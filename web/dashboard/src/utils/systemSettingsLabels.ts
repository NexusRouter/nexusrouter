/** Maps API `settings[].key` to i18n keys under `settings.fieldLabels.*`. */
export const SYSTEM_SETTINGS_LABEL_KEY: Record<string, string> = {
  http_listen_addr: 'httpListenAddr',
  upstream_timeout: 'upstreamTimeout',
  gateway_config_file: 'configFilePath',
  proxy_access_log_enabled: 'proxyLogEnabledRow',
  proxy_access_log_path: 'proxyLogPathRow',
  proxy_access_log_level: 'proxyLogLevelRow',
}

export function systemSettingsFieldLabelI18nKey(key: string): string | undefined {
  return SYSTEM_SETTINGS_LABEL_KEY[key]
}
