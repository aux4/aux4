const suggester = require('../lib/suggester');
const out = require('../lib/output');

describe('suggester', () => {
  let profile, commands;

  beforeEach(() => {
    out.println = jest.fn();
    profile = {};
    commands = [
      {
        value: 'cmd'
      }
    ];
    profile.commands = jest.fn().mockReturnValue(commands);
  });

  describe('suggest', () => {
    describe('with suggestion', () => {
      beforeEach(() => {
        suggester.suggest(profile, 'c');
      });

      it('prints the suggestion', () => {
        expect(out.println.mock.calls[0][0]).toEqual('What did you mean:');
        expect(out.println.mock.calls[1][0]).toEqual('  - ', 'cmd'.bold);
      });
    });

    describe('without suggestion', () => {
      beforeEach(() => {
        suggester.suggest(profile, 'x');
      });

      it('prints "command not found"', () => {
        expect(out.println).toHaveBeenCalledWith('Command not found: x');
      });
    });
  });
});
