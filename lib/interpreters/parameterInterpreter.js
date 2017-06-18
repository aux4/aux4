const variable = require('../variable');

module.exports = {
  interpret: function(command, action, args, parameters) {
    let result = action;

    let variables = variable.list(result);
    variables.forEach(function(name) {
      let value = parameters[name];
      if (value) {
        result = variable.replace(result, name, value);
      }
    });

    return result;
  }
};
