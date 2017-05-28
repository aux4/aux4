module.exports = {
  print: function() {
    let text = formatOutputText(arguments);
    process.stdout.write(text);
  },
  println: function() {
    let text = formatOutputText(arguments);
    text += '\n';
    process.stdout.write(text);
  }
};

function formatOutputText(args) {
  let text = '';
  for (let i = 0; i < args.length; i++) {
    let arg = args[i];
    if (arg === undefined || arg === '') {
      continue;
    }
    if (i > 0 && text.length > 0) {
      text += ' ';
    }
    text += arg;
  }
  return text;
}
