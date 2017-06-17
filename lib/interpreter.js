const interpreters = [];

module.exports = {
  add: function(interpreter) {
    interpreters.push(interpreter);
  },
  interpret: function(command, action, args, parameters) {
    let result = action;

    interpreters.forEach(function(interpreter) {
      result = interpreter.interpret(command, result, args, parameters);
    });

    return result;
  }
};
