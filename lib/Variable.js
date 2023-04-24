const PARAM_REGEX = /\${?(\w+)}?/g;
const VARIABLE_REGEX = /\${?(\w+)}?/;
const VARIABLE_WITH_INDEX_REGEX = /([^\[]+)\[([^\]]+)\]/;

const Variable = {
  list: function (action) {
    if (typeof action !== "string") {
      return [];
    }

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

            if (name.indexOf("[") > -1) {
              const index = name.replace(VARIABLE_WITH_INDEX_REGEX, "$2");
              name = name.replace(VARIABLE_WITH_INDEX_REGEX, "$1");
              result = result[name][index];
              return;
            }

            result = result[name];
          });
      } else if (variable.indexOf("[") > -1) {
        const index = variable.replace(VARIABLE_WITH_INDEX_REGEX, "$2");
        result = value[index];
      }

      if (result === undefined) {
        result = "";
      }

      if (responseAction === keyResult) {
        responseAction = result;
        return;
      }

      responseAction = responseAction.replace(keyResult, result);
    });
    return responseAction;
  }
};

module.exports = Variable;
