
const PARAM_REGEX = /\$\{?(\w+)\}?/g;
const VARIABLE_REGEX = /\$\{?(\w+)\}?/;

module.exports = {
  list: function(action) {
    let variables = {};
    let vars = action.match(PARAM_REGEX);
    if (vars) {
      vars.forEach(function(variable) {
        let key = variable.match(VARIABLE_REGEX)[1];
        variables[key] = true;
      });
    }
    return Object.keys(variables);
  },
  replace: function(action, key, value) {
    let responseAction = action;
    let regex = new RegExp(`\\$\\{?(${key}[^\\}]*)\\}?`, 'g');
    let results = responseAction.match(regex);
    results.forEach((keyResult) => {
      let variable = keyResult.replace(/\$\{?([^\}]+)\}?/, '$1');
      let result = value;

      if (variable.indexOf('.') > -1) {
        variable.split('.').splice(1).forEach((name) => {
          if (result === undefined) {
            return;
          }
          result = result[name];
        });
      }

      if (result === undefined) {
        result = '';
      }
      
      responseAction = responseAction.replace(keyResult, result);
    });
    return responseAction;
  }
};
