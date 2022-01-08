const Executor = require("../Executor");
const PREFIX = "profile:";

const ProfileExecutor = {
  with: function (config) {
    return function (interpreter, executorChain) {
      return {
        execute: function (command, action, args, parameters) {
          const executor = new Executor(config, executorChain);
          if (!action.startsWith(PREFIX)) {
            return false;
          }

          const profile = action.substring(PREFIX.length);
          executor.defineProfile(profile);
          executor.execute(args, parameters);

          return true;
        }
      };
    };
  }
};

module.exports = ProfileExecutor;
