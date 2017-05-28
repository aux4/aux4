const params = require('../lib/params.js');

describe('params', () => {
  describe('extract', () => {
    let args, obj;

    describe('when args is empty', () => {
      beforeEach(() => {
        args = [];

        obj = params.extract(args);
      });

      it('returns empty', () => {
        expect(obj).toEqual({});
      });
    });

    describe('when args is not empty', () => {
      beforeEach(() => {
        args = ['arg', '--single', '--name', 'the name', '--enabled'];

        obj = params.extract(args);
      });

      describe('size', () => {
        it('returns 3', () => {
          expect(Object.keys(obj).length).toEqual(3);
        });
      });

      describe('single', () => {
        it('returns true', () => {
          expect(obj.single).toBeTruthy();
        });
      });

      describe('name', () => {
        it('returns "the name"', () => {
          expect(obj.name).toEqual('the name');
        });
      });

      describe('enabled', () => {
        it('returns false', () => {
          expect(obj.enabled).toBeTruthy();
        });
      });
    });
  });
});
