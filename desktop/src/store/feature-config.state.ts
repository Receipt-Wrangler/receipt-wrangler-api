import { Injectable } from "@angular/core";
import { Action, createSelector, Selector, State, StateContext, } from "@ngxs/store";
import { FeatureConfig, OidcProviderSummary } from "../open-api";
import { SetFeatureConfig } from "./feature-config.state.actions";

@State<FeatureConfig>({
  name: "featureConfig",
  defaults: {
    enableLocalSignUp: true,
    aiPoweredReceipts: false,
    loginQrUrl: "",
    oidcProviders: [],
  },
})
@Injectable()
export class FeatureConfigState {
  @Selector()
  static enableLocalSignUp(state: FeatureConfig): boolean {
    return state.enableLocalSignUp as boolean;
  }

  @Selector()
  static aiPoweredReceipts(state: FeatureConfig): boolean {
    return state.aiPoweredReceipts as boolean;
  }

  @Selector()
  static loginQrUrl(state: FeatureConfig): string | undefined {
    return state.loginQrUrl;
  }

  @Selector()
  static oidcProviders(state: FeatureConfig): OidcProviderSummary[] {
    return state.oidcProviders ?? [];
  }

  @Selector()
  static featureConfig(state: FeatureConfig): FeatureConfig {
    return state;
  }

  static hasFeature(feature: string) {
    return createSelector([FeatureConfigState], (state: FeatureConfig) => {
      return !!(state as any)[feature];
    });
  }

  @Action(SetFeatureConfig)
  setFeatureConfig(
    { patchState }: StateContext<FeatureConfig>,
    payload: SetFeatureConfig
  ) {
    patchState({
      aiPoweredReceipts: payload.config?.aiPoweredReceipts,
      enableLocalSignUp: payload.config?.enableLocalSignUp,
      loginQrUrl: payload.config?.loginQrUrl,
      // Coalesced rather than assigned straight through: the field is optional on
      // the contract, so a server that predates it sends nothing, and consumers
      // should still get an array to iterate.
      oidcProviders: payload.config?.oidcProviders ?? [],
    });
  }
}
