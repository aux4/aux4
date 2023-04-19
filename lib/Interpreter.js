function Interpreter() {
  const interpreters = [];

  return {
    add: function (interpreter) {
      interpreters.push(interpreter);
    },
    interpret: async function (command, action, args, parameters) {
      let result = action;

      for (const interpreter of interpreters) {
        result = await interpreter.interpret(command, result, args, parameters);
      }

      return result;
    }
  };
}

module.exports = Interpreter;
