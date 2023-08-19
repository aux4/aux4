const { ConfigLoader, Config } = require("@aux4/config");

class ConfigParameterRetriever {
  constructor(config) {
    this.aux4Config = config;
    this.init = false;
  }

  static with(config) {
    return new ConfigParameterRetriever(config);
  }

  async lookup(command, parameters, name) {
    if (!this.init) {
      this.config = getConfig(this.aux4Config);
      this.init = true;
    }

    if (this.config === undefined) return undefined;

    const configPath = await parameters.config;
    const config = this.config.get(configPath);

    if (config === undefined) return undefined;

    return config[name];
  }
}

function getConfig(config) {
  if (config) {
    const aux4Config = config.get();
    if (aux4Config && aux4Config.config) {
      return new Config(aux4Config.config);
    }
  }
  return ConfigLoader.load();
}

module.exports = ConfigParameterRetriever;
