const Crypto = require("../../lib/Crypto");

const CryptoInterpreter = require("../../lib/interpreters/CryptoInterpreter");
const cryptoInterpreter = new CryptoInterpreter();

describe("cryptoInterpreters", () => {
  describe("intercept", () => {
    let result;

    describe("without cryto", () => {
      beforeEach(() => {
        result = cryptoInterpreter.interpret({}, "mkdir test", [], {});
      });

      it("does not replace the action", () => {
        expect(result).toEqual("mkdir test");
      });
    });

    describe("with a single cryto", () => {
      beforeEach(() => {
        Crypto.decrypt = jest.fn(() => "1234");
        result = cryptoInterpreter.interpret({}, "connect -u root -p crypto(abcd)", [], {});
      });

      it("calls crypto decrypt", () => {
        expect(Crypto.decrypt).toHaveBeenCalledWith("abcd");
      });

      it("does not replace the action", () => {
        expect(result).toEqual("connect -u root -p 1234");
      });
    });

    describe("with multiple crytos", () => {
      beforeEach(() => {
        Crypto.decrypt = jest.fn().mockReturnValueOnce("1234").mockReturnValue("4321");
        result = cryptoInterpreter.interpret({}, "connect -u root -p crypto(abcd) -token crypto(dcba)", [], {});
      });

      it("calls crypto decrypt", () => {
        expect(Crypto.decrypt.mock.calls.length).toEqual(2);
        expect(Crypto.decrypt.mock.calls[0][0]).toEqual("abcd");
        expect(Crypto.decrypt.mock.calls[1][0]).toEqual("dcba");
      });

      it("does not replace the action", () => {
        expect(result).toEqual("connect -u root -p 1234 -token 4321");
      });
    });
  });
});
