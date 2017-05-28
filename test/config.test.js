const colors = require('colors');
const fs = require('fs');

const config = require('../lib/config');

const out = require('../lib/output');

describe('config', () => {
  beforeEach(() => {
    out.println = jest.fn();
  });

  describe('file', () => {
    it('returns an empty object without profiles', () => {
      expect(config.file()).toEqual({profiles: []});
    });
  });

  describe('load config file', () => {
    describe('when there is no config file', () => {
      beforeEach(() => {
        fs.access = jest.fn((name, cb) => cb('error'));

        config.load();
      });

      it('prints ".aux4 file not found"', () => {
        expect(out.println).toHaveBeenCalledWith('.aux4 file not found'.red);
      });
    });

    describe('when there is config file', () => {
      describe('with error to read', () => {
        beforeEach(() => {
          fs.access = jest.fn((name, cb) => cb());
          fs.readFile = jest.fn((name, cb) => cb('error'));

          config.load();
        });

        it('does not print ".aux4 file not found"', () => {
          expect(out.println).not.toHaveBeenCalledWith('.aux4 file not found'.red);
        });

        it('prints "error reading .aux4 file, check the permissions"', () => {
          expect(out.println).toHaveBeenCalledWith(
            'error reading .aux4 file, check the permissions'.red
          );
        });
      });

      describe('without error to read', () => {
        describe('with error to parse', () => {
          let configFile;

          beforeEach(() => {
            configFile = 'wrong json format';

            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb(undefined, configFile));

            config.load();
          });

          it('does not print ".aux4 file not found"', () => {
            expect(out.println).not.toHaveBeenCalledWith('.aux4 file not found'.red);
          });

          it('does not print "error reading .aux4 file, check the permissions"', () => {
            expect(out.println).not.toHaveBeenCalledWith(
              'error reading .aux file, check the permissions'.red
            );
          });

          it('prints ".aux4 is not a valid json file"', () => {
            expect(out.println).toHaveBeenCalledWith('.aux4 is not a valid json file'.red);
          });
        });

        describe('without error to parse', () => {
          let configFile;

          beforeEach(() => {
            configFile = {
              profiles: []
            };

            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb(undefined, JSON.stringify(configFile)));

            config.load();
          });

          it('does not print ".aux4 file not found"', () => {
            expect(out.println).not.toHaveBeenCalledWith('.aux4 file not found'.red);
          });

          it('does not print "error reading .aux4 file, check the permissions"', () => {
            expect(out.println).not.toHaveBeenCalledWith(
              'error reading .aux4 file, check the permissions'.red
            );
          });

          it('does not print ".aux4 is not a valid json file"', () => {
            expect(out.println).not.toHaveBeenCalledWith('.aux4 is not a valid json file'.red);
          });

          describe('get config file', () => {
            it('returns the object parsed from json', () => {
              expect(config.file()).toEqual(configFile);
            });
          });
        });
      });
    });
  });
});
