module.exports = {
  extract: function(args) {
    let params = {};

    for (let i = 0; i < args.length; i++) {
      let arg = args[i];
      let name, value;
      if (arg.startsWith('--')) {
        name = arg.substring(2);
        value = true;
        if (i + 1 < args.length) {
          if (!args[i + 1].startsWith('--')) {
            value = args[i + 1];
            i++;
          }
        }
        params[name] = value;
      }
    }

    return params;
  }
};
