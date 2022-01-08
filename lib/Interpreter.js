function Interpreter() {
  const interpreters = [];

  return {
    add: function (interpreter) {
      interpreters.push(interpreter);
    },
    interpret: function (command, action, args, parameters) {
      let result = action;

      interpreters.forEach(interpreter => {
        result = interpreter.interpret(command, result, args, parameters);
      });

      return result;
    }
  };
}

module.exports = Interpreter;
