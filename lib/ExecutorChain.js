const colors = require("colors");
const out = require("./Output");

function ExecutorChain(interpreter) {
  const executors = [];

  return {
    register: function (executor) {
      executors.push(new executor(interpreter, this));
    },
    execute: function (command, args, parameters) {
      if (command.execute === undefined) {
        out.println("execute is not defined".red);
        return;
      }

      if (typeof command.execute === "function") {
        command.execute(parameters, args, command);
        return;
      }

      const actions = command.execute;

      for (let x = 0; x < actions.length; x++) {
        const action = actions[x];
        for (let i = 0; i < executors.length; i++) {
          const executor = executors[i];
          let response;

          try {
            response = executor.execute(command, action, args, parameters);
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
