const iterator = require('../lib/iterator.js');

describe('iterator', () => {
  describe('when the array is empty', () => {
    let emptyArrayIterator;

    beforeEach(() => {
      emptyArrayIterator = iterator([]);
    });

    describe('hasNext', () => {
      it('should be false', () => {
        expect(emptyArrayIterator.hasNext()).toBeFalsy();
      });
    });

    describe('next', () => {
      it('should throw an error', () => {
        expect(emptyArrayIterator.next()).toThrow();
      });
    });
  });

  describe('when the array is not empty', () => {
    let arrayIterator;

    beforeEach(() => {
      arrayIterator = iterator([1, 2]);
    });

    describe('hasNext', () => {
      it('should be true', () => {
        expect(arrayIterator.hasNext()).toBeTruthy();
      });
    });

    describe('next', () => {
      it('should throw an error', () => {
        expect(arrayIterator.next()).toEqual(1);
      });
    });
  });
});
