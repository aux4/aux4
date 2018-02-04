const colors = require('colors');
const fs = require('fs');

const config = require('../lib/config');

const out = require('../lib/output');

describe('config', () => {
  let profile$;

  beforeEach(() => {
    profile$ = {
      name: 'aux4',
      commands: [
        {
          value: 'encrypt',
          execute: ['crypto:encrypt'],
          help: {
            description: 'Encrypt value.\nTo make the encryption more safe, you can define a special key in the environment variable AUX4_SECURITY_KEY.'
          }
        },
        {
          value: 'upgrade',
          execute: ['npm install --global aux4'],
          help: {
            description: 'Upgrade the aux4 version.'
          }
        }
      ]
    };

    out.println = jest.fn();
  });

  describe('get', () => {
    it('returns an empty object without profiles', () => {
      expect(config.get()).toEqual({ profiles: [profile$] });
    });
  });

  describe('load config file', () => {
    describe('when there is no config file', () => {
      let callback;

      beforeEach(() => {
        fs.access = jest.fn((name, cb) => cb('error'));

        callback = jest.fn();

        config.loadFile(undefined, callback);
      });

      it('prints ".aux4 file not found"', () => {
        expect(out.println).toHaveBeenCalledWith('.aux4 file not found'.red);
      });

      it('calls the callback with error', () => {
        expect(callback).toHaveBeenCalledWith(new Error('.aux4 file not found'));
      });
    });

    describe('when there is config file', () => {
      describe('with default file name', () => {
        describe('with error to read', () => {
          let callback;

          beforeEach(() => {
            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb('error'));

            callback = jest.fn();

            config.loadFile(undefined, callback);
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
            expect(callback).toHaveBeenCalledWith(
              new Error('error reading .aux4 file, check the permissions')
            );
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

              config.loadFile(undefined, callback);
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

              config.loadFile(undefined, callback);
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
                expect(config.get()).toEqual({ profiles: [profile$] });
              });
            });
          });
        });
      });

      describe('with custom file name', () => {
        let fileName;

        beforeEach(() => {
          fileName = 'newFile.aux4';
        });

        describe('with error to read', () => {
          let callback;

          beforeEach(() => {
            fs.access = jest.fn((name, cb) => cb());
            fs.readFile = jest.fn((name, cb) => cb('error'));

            callback = jest.fn();

            config.loadFile(fileName, callback);
          });

          it('does not print "newFile.aux4 file not found"', () => {
            expect(out.println).not.toHaveBeenCalledWith('newFile.aux4 file not found'.red);
          });

          it('prints "error reading newFile.aux4 file, check the permissions"', () => {
            expect(out.println).toHaveBeenCalledWith(
              'error reading newFile.aux4 file, check the permissions'.red
            );
          });

          it('calls the callback with error', () => {
            expect(callback).toHaveBeenCalledWith(
              new Error('error reading newFile.aux4 file, check the permissions')
            );
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

              config.loadFile(fileName, callback);
            });

            it('does not print "newFile.aux4 file not found"', () => {
              expect(out.println).not.toHaveBeenCalledWith('newFile.aux4 file not found'.red);
            });

            it('does not print "error reading newFile.aux4 file, check the permissions"', () => {
              expect(out.println).not.toHaveBeenCalledWith(
                'error reading newFile.aux file, check the permissions'.red
              );
            });

            it('prints "newFile.aux4 is not a valid json file"', () => {
              expect(out.println).toHaveBeenCalledWith('newFile.aux4 is not a valid json file'.red);
            });

            it('calls the callback with error', () => {
              expect(callback).toHaveBeenCalledWith(
                new Error('newFile.aux4 is not a valid json file')
              );
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

              config.loadFile(fileName, callback);
            });

            it('does not print "newFile.aux4 file not found"', () => {
              expect(out.println).not.toHaveBeenCalledWith('newFile.aux4 file not found'.red);
            });

            it('does not print "error reading newFile.aux4 file, check the permissions"', () => {
              expect(out.println).not.toHaveBeenCalledWith(
                'error reading newFile.aux4 file, check the permissions'.red
              );
            });

            it('does not print "newFile.aux4 is not a valid json file"', () => {
              expect(out.println).not.toHaveBeenCalledWith(
                'newFile.aux4 is not a valid json file'.red
              );
            });

            it('calls the callback without error', () => {
              expect(callback).toHaveBeenCalledWith(undefined);
            });

            describe('get config file', () => {
              it('returns the object parsed from json', () => {
                expect(config.get()).toEqual({ profiles: [profile$] });
              });
            });
          });
        });
      });

      describe('override configuration', () => {
        let configFile, configFileA, configFileB, callback;

        beforeEach(() => {
          configFileA = {
            profiles: [
              {
                name: 'A',
                commands: [{ value: 'one', execute: ['oneA'] }, { value: 'two', execute: ['twoA'] }]
              },
              {
                name: 'B',
                commands: [
                  { value: 'three', execute: ['threeB'] },
                  { value: 'five', execute: ['fiveB'] }
                ]
              }
            ]
          };

          configFileB = {
            profiles: [
              { name: 'A', commands: [{ value: 'four', execute: ['fourA'] }] },
              { name: 'B', commands: [{ value: 'three', execute: ['3rd'] }] }
            ]
          };

          configFile = {
            profiles: [
              profile$,
              {
                name: 'A',
                commands: [
                  { value: 'one', execute: ['oneA'] },
                  { value: 'two', execute: ['twoA'] },
                  { value: 'four', execute: ['fourA'] }
                ]
              },
              {
                name: 'B',
                commands: [
                  { value: 'three', execute: ['3rd'] },
                  { value: 'five', execute: ['fiveB'] }
                ]
              }
            ]
          };

          callback = jest.fn();

          fs.access = jest.fn((name, cb) => cb());
          fs.readFile = jest.fn((name, cb) => cb(undefined, JSON.stringify(configFileA)));

          config.loadFile('a.aux4', callback);

          config.load(configFileB, callback);
        });

        describe('get config file', () => {
          it('returns the object parsed from json', () => {
            expect(config.get()).toEqual(configFile);
          });
        });
      });
    });
  });
});
