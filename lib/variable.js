
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
    let regex = new RegExp(`\\$\\{?${key}\\}?`, 'g');
    return action.replace(regex, value);
  }
};
