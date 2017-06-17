const variable = require('../lib/variable');

describe('variable', () => {
  describe('list', () => {
    describe('without variables', () => {
      let list, action;

      beforeEach(() => {
        action = 'mkdir folder';

        list = variable.list(action);
      });

      it('returns an empty list', () => {
        expect(list).toEqual([]);
      });
    });

    describe('single variable', () => {
      let list, action;

      beforeEach(() => {
        action = 'mkdir ${folder}';

        list = variable.list(action);
      });

      it('returns folder in the list', () => {
        expect(list).toEqual(['folder']);
      });
    });

    describe('duplicated variable', () => {
      let list, action;

      beforeEach(() => {
        action = 'echo ${name} / $name';

        list = variable.list(action);
      });

      it('returns name in the list', () => {
        expect(list).toEqual(['name']);
      });
    });

    describe('multiple variables', () => {
      let list, action;

      beforeEach(() => {
        action = 'echo ${firstName} ${middleName} $lastName';

        list = variable.list(action);
      });

      it('returns first, middle and last name in the list', () => {
        expect(list).toEqual(['firstName', 'middleName', 'lastName']);
      });
    });
  });

  describe('replace', () => {
    describe('single variable', () => {
      let result, action, key, value;

      beforeEach(() => {
        action = 'mkdir ${folder}';
        key = 'folder';
        value = 'test';

      	result = variable.replace(action, key, value);
      });

      it('replaces variable', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('duplicated variable', () => {
      let result, action, key, value;

      beforeEach(() => {
        action = 'echo ${name} / $name';
        key = 'name';
        value = 'test';

      	result = variable.replace(action, key, value);
      });

      it('replaces variables', () => {
        expect(result).toEqual('echo test / test');
      });
    });

    describe('multiple variables', () => {
      let result, action, key, value;

      beforeEach(() => {
        action = 'echo ${firstName} ${lastName}';
        key = 'firstName';
        value = 'John';

      	result = variable.replace(action, key, value);
      });

      it('replaces $firstName', () => {
        expect(result).toEqual('echo John ${lastName}');
      });
    });
  });
});
