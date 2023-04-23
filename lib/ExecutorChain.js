const colors = require("colors");
const out = require("./Output");

function ExecutorChain(interpreter) {
  const executors = [];

  return {
    register: function (executor) {
      executors.push(new executor(interpreter, this));
    },
    execute: async function (command, args, parameters) {
      if (command.execute === undefined) {
        out.println("execute is not defined".red);
        return;
      }

      if (typeof command.execute === "function") {
        command.execute(parameters, args, command);
        return;
      }

      const actions = command.execute;

      for (const action of actions) {
        for (const executor of executors) {
          let response;

          try {
            response = await executor.execute(command, action, args, parameters);
          } catch (e) {
            return;
          }

          if (response) {
            break;
          }
        }
      }
    }
  };
}

module.exports = ExecutorChain;
