const config = require('../lib/config');
const Profile = require('../lib/profile');
const executorChain = require('../lib/executorChain');

const executor = require('../lib/executor');

describe('executor', () => {
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

  describe('initialize executor', () => {
    beforeEach(() => {
    	executor.init();
    });

    it('calls config file', () => {
      expect(config.file).toHaveBeenCalled();
    });

    it('creates firstProfile', () => {
      expect(executor.profile('firstProfile').name()).toEqual('firstProfile');
    });

    it('creates secondProfile', () => {
      expect(executor.profile('secondProfile').name()).toEqual('secondProfile');
    });

    describe('current profile', () => {
      it('returns "main"', () => {
        expect(executor.currentProfile()).toEqual('main');
      });
    });
  });

  describe('change current profile', () => {
    describe('when profile does not exists', () => {
      it('throw an error', () => {
        expect(() => {
          executor.defineProfile('abc');
        }).toThrow(new Error('profile abc does not exists'));
      });
    });

    describe('when profile exists', () => {
      beforeEach(() => {
        executor.defineProfile('firstProfile');
      });

      describe('current profile', () => {
        it('returns "firstProfile"', () => {
          expect(executor.currentProfile()).toEqual('firstProfile');
        });
      });
    });
  });

  describe('execute', () => {
    beforeEach(() => {
      executorChain.execute = jest.fn();

      executor.init();
    	executor.defineProfile('firstProfile');
      executor.execute(['cmd', '--enable', 'true']);
    });

    it('calls executorChain', () => {
      expect(executorChain.execute).toHaveBeenCalledWith(configProfiles.profiles[0].commands[0], ['--enable', 'true']);
    });
  });
});
