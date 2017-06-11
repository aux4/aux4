const params = require('./params');

module.exports = {
  interpret: function(command, args, parameters) {
    let result = command;

    let keys = Object.keys(parameters);
    keys.forEach(function(key) {
      let regex = new RegExp(`\\$(\\{)?${key}(\\})?`, 'g');
      result = result.replace(regex, parameters[key]);
    });

    return result;
  }
};
