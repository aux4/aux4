module.exports = function(array) {
  const length = array.length;
  let index = -1;

  return {
    hasNext: function() {
      return index + 1 < length;
    },
    next: function() {
      return array[++index];
    }
  };
};
