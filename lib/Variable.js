const PARAM_REGEX = /\${?(\w+)}?/g;
const VARIABLE_REGEX = /\${?(\w+)}?/;

const Variable = {
  list: function (action) {
    const variables = {};
    const vars = action.match(PARAM_REGEX);
    if (vars) {
      vars.forEach(function (variable) {
        const key = variable.match(VARIABLE_REGEX)[1];
        variables[key] = true;
      });
    }
    return Object.keys(variables);
  },
  replace: function (action, key, value) {
    let responseAction = action;
    const regex = new RegExp(`\\$\{?(${key}[^}\\s]*)}?`, "g");
    const results = responseAction.match(regex);
    results.forEach(keyResult => {
      const variable = keyResult.replace(/\${?([^}\s]+)}?/, "$1");
      let result = value;

      if (variable.indexOf(".") > -1) {
        variable
          .split(".")
          .splice(1)
          .forEach(name => {
            if (result === undefined) {
              return;
            }
            result = result[name];
          });
      }

      if (result === undefined) {
        result = "";
      }

      responseAction = responseAction.replace(keyResult, result);
    });
    return responseAction;
  }
};

module.exports = Variable;
