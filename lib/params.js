module.exports = {
  extract: function(args) {
    const params = {};
    const itemsToRemove = [];

    for (let i = 0; i < args.length; i++) {
      const arg = args[i];

      if (arg.startsWith('--')) {
        itemsToRemove.push(i);

        const name = arg.substring(2);
        let value = true;
        if (i + 1 < args.length) {
          if (!args[i + 1].startsWith('--')) {
            value = args[i + 1];
            itemsToRemove.push(i + 1);
            i++;
          }
        }
        params[name] = value;
      }
    }

    itemsToRemove.reverse().forEach(function(index){
      args.splice(index, 1);
    });

    return params;
  }
};
