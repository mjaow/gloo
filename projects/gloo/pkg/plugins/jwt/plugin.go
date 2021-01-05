package jwt

import (
	"fmt"
	"time"

	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoyjwt "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/enterprise/options/jwt"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins"
	"google.golang.org/protobuf/types/known/durationpb"
)

const filterName = "envoy.filters.http.jwt_authn"

var pluginStage = plugins.DuringStage(plugins.AuthNStage)

var _ plugins.Plugin = &Plugin{}
var _ plugins.HttpFilterPlugin = &Plugin{}

func NewJwtPlugin() *Plugin {
	return &Plugin{}
}

type Plugin struct {
}

func (p *Plugin) Init(params plugins.InitParams) error {
	return nil
}

func (p *Plugin) HttpFilters(params plugins.Params, listener *v1.HttpListener) ([]plugins.StagedHttpFilter, error) {
	providers := map[string]*jwt.Provider{}

	for _, host := range listener.GetVirtualHosts() {
		for k, v := range host.GetOptions().GetJwt().GetProviders() {
			if _, exists := providers[k]; exists {
				return nil, fmt.Errorf("jwt providers conflict.Already exists jwt provider with name [%s]", k)
			}

			providers[k] = v
		}
	}

	cfg, err := createJwtProviders(providers)

	if err != nil {
		return nil, err
	}

	filter, err := plugins.NewStagedFilterWithConfig(filterName, cfg, pluginStage)

	if err != nil {
		return nil, err
	}

	return []plugins.StagedHttpFilter{filter}, nil
}

func createJwtProviders(providers map[string]*jwt.Provider) (*envoyjwt.JwtAuthentication, error) {
	var providersMap = map[string]*envoyjwt.JwtProvider{}
	var rules []*envoyjwt.RequirementRule

	for k, p := range providers {
		if provider, err := convertProvider(p); err != nil {
			return nil, fmt.Errorf("provider %s with error %v", k, err)
		} else {
			providersMap[k] = provider

			rules = append(rules, &envoyjwt.RequirementRule{
				Match: &envoy_config_route_v3.RouteMatch{
					PathSpecifier: &envoy_config_route_v3.RouteMatch_Prefix{
						Prefix: "/",
					},
				},

				RequirementType: &envoyjwt.RequirementRule_Requires{
					Requires: &envoyjwt.JwtRequirement{
						RequiresType: &envoyjwt.JwtRequirement_ProviderName{
							ProviderName: k,
						},
					},
				},
			})
		}
	}

	config := &envoyjwt.JwtAuthentication{
		Providers: providersMap,
		Rules:     rules,
	}

	return config, nil
}

func convertProvider(p *jwt.Provider) (*envoyjwt.JwtProvider, error) {
	if p == nil {
		return nil, fmt.Errorf("must not be nil provider")
	}

	local := p.GetJwks().GetLocal()
	remote := p.GetJwks().GetRemote()

	provider := &envoyjwt.JwtProvider{
		Issuer:    p.Issuer,
		Audiences: p.Audiences,
	}

	if local != nil {
		provider.JwksSourceSpecifier = &envoyjwt.JwtProvider_LocalJwks{
			LocalJwks: &v3.DataSource{
				Specifier: &v3.DataSource_InlineString{
					InlineString: local.GetKey(),
				},
			},
		}
		return provider, nil
	} else if remote != nil {
		provider.JwksSourceSpecifier = &envoyjwt.JwtProvider_RemoteJwks{
			RemoteJwks: &envoyjwt.RemoteJwks{
				HttpUri: &v3.HttpUri{
					Uri: remote.GetUrl(),
					HttpUpstreamType: &v3.HttpUri_Cluster{
						Cluster: fmt.Sprintf("%s_%s", remote.GetUpstreamRef().GetName(), remote.GetUpstreamRef().GetNamespace()),
					},
					Timeout: durationpb.New(time.Second * 300),
				},
				CacheDuration: remote.GetCacheDuration(),
			},
		}
		return provider, nil
	} else {
		return nil, fmt.Errorf("either remote or local must be not-nil")
	}
}
