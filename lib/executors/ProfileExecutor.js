const Executor = require("../Executor");
const PREFIX = "profile:";

const ProfileExecutor = {
  with: function (config) {
    return function (interpreter, executorChain) {
      return {
        execute: async function (command, action, args, parameters) {
          if (!action.startsWith(PREFIX)) {
            return false;
          }

          const executor = new Executor(config, executorChain);
          const profile = action.substring(PREFIX.length);
          executor.defineProfile(profile);
          await executor.execute(args, parameters);

          return true;
        }
      };
    };
  }
};

module.exports = ProfileExecutor;
