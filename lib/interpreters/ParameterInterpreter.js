const Variable = require("../Variable");

function ParameterInterpreter() {
  return {
    interpret: async function (command, action, args, parameters) {
      let result = action;

      const variables = Variable.list(result);
      variables.forEach(function (name) {
        const value = parameters[name];
        if (value) {
          result = Variable.replace(result, name, value);
        }
      });

      return result;
    }
  };
}

module.exports = ParameterInterpreter;
