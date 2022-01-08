describe("crypto", () => {
  let Crypto, text, encryptedText;

  describe("without env variable", () => {
    beforeEach(() => {
      jest.resetModules();
      Crypto = require("../lib/Crypto");

      text = "this is the pure text";
      encryptedText = Crypto.encrypt(text);
    });

    describe("text is encrypted", () => {
      it("returns the text encrypted", () => {
        expect(encryptedText).not.toEqual(text);
      });
    });

    describe("decrypt", () => {
      it("returns the original text", () => {
        expect(Crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });

  describe("with env variable", () => {
    beforeEach(() => {
      jest.resetModules();
      process.env.AUX4_SECURITY_KEY = "DF62446FD8C45959";
      Crypto = require("../lib/Crypto");

      text = "this is the pure text";
      encryptedText = Crypto.encrypt(text);
    });

    describe("text is encrypted", () => {
      it("returns the text encrypted", () => {
        expect(encryptedText).not.toEqual(text);
      });
    });

    describe("decrypt", () => {
      it("returns the original text", () => {
        expect(Crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });
});
