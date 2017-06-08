const Profile = require('../lib/profile');

const config = require('../lib/config');

describe('profile', () => {
  let configProfiles;
  beforeEach(() => {
    configProfiles = {
      profiles: [
        {
          name: 'firstProfile',
          commands: [
            {
              value: 'cmd',
              execute: ['mkdir first', 'cd first']
            }
          ]
        },
        {
          name: 'secondProfile',
          commands: [
            {
              value: 'cmd',
              execute: ['mkdir second', 'cd second']
            }
          ]
        }
      ]
    };

    config.file = jest.fn().mockReturnValue(configProfiles);
  });

  describe('create firstProfile', () => {
    let firstProfile, firstProfileName;
    beforeEach(() => {
      firstProfileName = 'firstProfile';
      firstProfile = new Profile(firstProfileName);
    });

    describe('firstProfile name', () => {
      it('returns "firstProfile"', () => {
        expect(firstProfile.name()).toEqual(firstProfileName);
      });
    });

    describe('firstProfile commands', () => {
      it('returns firstProfile commands', () => {
        expect(firstProfile.commands()).toBe(configProfiles.profiles[0].commands);
      });
    });

    describe('get "cmd" command', () => {
      it('returns cmd command', () => {
        expect(firstProfile.command('cmd')).toBe(configProfiles.profiles[0].commands[0]);
      });
    });

    describe('get non-existent command', () => {
      it('returns undefined', () => {
        expect(firstProfile.command('x')).toBeUndefined();
      });
    });
  });

  describe('create secondProfile', () => {
    let secondProfile, secondProfileName;
    beforeEach(() => {
      secondProfileName = 'secondProfile';
    	secondProfile = new Profile(secondProfileName);
    });

    describe('secondProfile name', () => {
      it('returns "secondProfile"', () => {
        expect(secondProfile.name()).toEqual(secondProfileName);
      });
    });

    describe('secondProfile commands', () => {
      it('returns secondProfile commands', () => {
        expect(secondProfile.commands()).toBe(configProfiles.profiles[1].commands);
      });
    });
  });

  describe('create profile which is not on the config properties', () => {
    describe('thirdProfile', () => {
      it('returns undefined', () => {
        expect(() => {
          new Profile('thirdProfile');
        }).toThrow('profile thirdProfile not found in the configuration file');
      });
    });
  });
});
