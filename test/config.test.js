const colors = require('colors');
const fs = require('fs');

const config = require('../lib/config');

const out = require('../lib/output');

describe('config', () => {
  beforeEach(() => {
    out.println = jest.fn();
  });

  describe('$', () => {
    let profile;

    beforeEach(() => {
    	profile = config.$();
    });

    describe('name', () => {
      it('returns $', () => {
        expect(profile.name).toEqual('$');
      });
    });

    describe('commands', () => {
      it('should have one command', () => {
        expect(profile.commands.length).toEqual(2);
      });

      it('returns encrypt command', () => {
        expect(profile.commands[0].value).toEqual('encrypt');
      });

      it('returns encrypt execute', () => {
        expect(profile.commands[0].execute).toEqual(['crypto:encrypt']);
      });

      it('returns upgrade command', () => {
        expect(profile.commands[1].value).toEqual('upgrade');
      });

      it('returns upgrade execute', () => {
        expect(profile.commands[1].execute).toEqual(['npm install --global aux4']);
      });
    });
  });

  describe('file', () => {
    it('returns an empty object without profiles', () => {
      expect(config.file()).toEqual({profiles: []});
    });
  });

  describe('load config file', () => {
    describe('when there is no config file', () => {
      let callback;

      beforeEach(() => {
        fs.access = jest.fn((name, cb) => cb('error'));

        callback = jest.fn();

        config.load(callback);
      });

      it('prints ".aux4 file not found"', () => {
        expect(out.println).toHaveBeenCalledWith('.aux4 file not found'.red);
      });

      it('calls the callback with error', () => {
        expect(callback).toHaveBeenCalledWith(new Error('.aux4 file not found'));
      });
    });

    describe('when there is config file', () => {
      describe('with error to read', () => {
        let callback;

        beforeEach(() => {
          fs.access = jest.fn((name, cb) => cb());
          fs.readFile = jest.fn((name, cb) => cb('error'));

          callback = jest.fn();

          config.load(callback);
        });

        it('does not print ".aux4 file not found"', () => {
          expect(out.println).not.toHaveBeenCalledWith('.aux4 file not found'.red);
        });

        it('prints "error reading .aux4 file, check the permissions"', () => {
          expect(out.println).toHaveBeenCalledWith(
            'error reading .aux4 file, check the permissions'.red
          );
        });

        it('calls the callback with error', () => {
          expect(callback).toHaveBeenCalledWith(new Error('error reading .aux4 file, check the permissions'));
        });
      });

      describe('without error to read', () => {
        describe('with error to parse', () => {
          let configFile, callback;

          beforeEach(() => {
            configFile = 'wrong json format';

            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb(undefined, configFile));

            callback = jest.fn();

            config.load(callback);
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

          it('calls the callback with error', () => {
            expect(callback).toHaveBeenCalledWith(new Error('.aux4 is not a valid json file'));
          });
        });

        describe('without error to parse', () => {
          let configFile, callback;

          beforeEach(() => {
            configFile = {
              profiles: []
            };

            callback = jest.fn();

            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb(undefined, JSON.stringify(configFile)));

            config.load(callback);
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

          it('calls the callback without error', () => {
            expect(callback).toHaveBeenCalledWith(undefined);
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
