const Config = require("./Config");
const ExecutorChain = require("./ExecutorChain");
const Executor = require("./Executor");
const ProfileExecutor = require("./executors/ProfileExecutor");
const Parameters = require("./Parameters");
const Suggester = require("./Suggester");

function defaultOptions() {
  const config = new Config();
  const executorChain = defaultExecutorChain(config);

  return {
    config,
    executorChain,
    suggester: Suggester,
    aux4: { profiles: [] }
  };
}

function defaultExecutorChain(config) {
  const executorChain = new ExecutorChain();
  executorChain.register(ProfileExecutor.with(config));
  return executorChain;
}

const DEFAULT_OPTIONS = defaultOptions();

const Engine = function (options = {}) {
  const config = options.config || DEFAULT_OPTIONS.config;
  const executorChain =
    options.executorChain || (options.config ? defaultExecutorChain(options.config) : DEFAULT_OPTIONS.executorChain);
  const suggester = options.suggester || DEFAULT_OPTIONS.suggester;
  const aux4 = options.aux4 || DEFAULT_OPTIONS.aux4;

  config.load(aux4);

  const executor = new Executor(config, executorChain, suggester);

  return {
    run: function (args) {
      const parameters = Parameters.extract(args);
      executor.execute(args, parameters);
    }
  };
};

module.exports = Engine;
