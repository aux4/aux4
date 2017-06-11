const profileExecutor = require('../../lib/executors/profileExecutor');

const executor = require('../../lib/executor');

describe('profileExecutor', () => {
  beforeEach(() => {
  	executor.defineProfile = jest.fn();
  	executor.execute = jest.fn();
  });

  describe('execute', () => {
    let action, args, parameters, result;

    describe('when is not a profile', () => {
      beforeEach(() => {
      	action = 'mkdir test';
        args = [];
        parameters = {};

        result = profileExecutor.execute({}, action, args, parameters);
      });

      it('returns false', () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when is a profile', () => {
      let profile;

      beforeEach(() => {
        profile = 'git';
      	action = 'profile:' + profile;
        args = ['push'];
        parameters = {};

        result = profileExecutor.execute({}, action, args, parameters);
      });

      it('calls "executor.defineProfile" with the profile', () => {
        expect(executor.defineProfile).toHaveBeenCalledWith(profile);
      });

      it('executes "executor.execute"', () => {
        expect(executor.execute).toHaveBeenCalledWith(args, parameters);
      });

      it('returns true', () => {
        expect(result).toBeTruthy();
      });
    });
  });
});
