const colors = require("colors");

const Crypto = require("../../lib/Crypto");
const {Output:out} = require("@aux4/engine");

const EncryptExecutor = require("../../lib/executors/EncryptExecutor");
const encryptExecutor = new EncryptExecutor();

describe("encryptExecutor", () => {
  let command, action, args, parameters, result, spyOnEncrypt;

  beforeEach(() => {
    command = {};
    args = [];
    parameters = {};

    Crypto.encrypt = jest.fn(() => "****");
    out.println = jest.fn();
  });

  describe("execute", () => {
    describe('when the prefix is not "crypto:encrypt"', () => {
      beforeEach(async () => {
        action = "log:test";
        result = await encryptExecutor.execute(command, action, args, parameters);
      });

      it("returns false", () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when prefix is "crypto:encrypt"', () => {
      describe("args is empty", () => {
        beforeEach(async () => {
          action = "crypto:encrypt";
          result = await encryptExecutor.execute(command, action, args, parameters);
        });

        it('prints "There is nothing to encrypt" message', () => {
          expect(out.println).toHaveBeenCalledWith("There is nothing to encrypt".red);
        });

        it("returns true", () => {
          expect(result).toBeTruthy();
        });
      });

      describe("args is not empty", () => {
        beforeEach(async () => {
          action = "crypto:encrypt";
          args = ["abcd"];
          result = await encryptExecutor.execute(command, action, args, parameters);
        });

        it('does not prints "There is nothing to encrypt" message', () => {
          expect(out.println).not.toHaveBeenCalledWith("There is nothing to encrypt".red);
        });

        it("calls crypto.encrypt", () => {
          expect(Crypto.encrypt).toHaveBeenCalledWith(args[0]);
        });

        it("prints encrypted text", () => {
          expect(out.println.mock.calls.length).toEqual(1);
          expect(out.println).toHaveBeenCalledWith("****");
        });

        it("returns true", () => {
          expect(result).toBeTruthy();
        });
      });
    });
  });
});
