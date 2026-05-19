describe("E2E tests: registration flow", () => {
  beforeEach(() => {
    cy.task('db:clean');
  });

  //test 1
  it("should register user, create business and verify the code", () => {
    const testEmail = `reg_${Date.now()}@example.com`;
    const testPassword = '123456TEST';
    const testBusinessName = `Test Business ${Date.now()}`;
    const testBusinessType = 'other'

    cy.visit('/register');
    fillForm(testEmail, testPassword, testBusinessName);
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/verify');
    const code = '123456';
    enterVerificationCode(code);
    cy.get('[data-cy="verify-submit"]').click();
    cy.url().should('include', '/login');

    cy.task('db:findUserByEmail', testEmail).then((user) => {
      expect(user).to.not.be.null;
      expect(user.login).to.equal(testEmail);
      expect(user.business_id).to.be.greaterThan(0);
    });

    cy.task('db:findBusinessByName', testBusinessName).then((business) => {
      expect(business).to.not.be.null;
      expect(business.name).to.equal(testBusinessName);
      expect(business.type).to.equal(testBusinessType);
    });
  });

  //test2
  it('should show error with existing email', () => {
    const testEmail = `dup_${Date.now()}@example.com`;
    const testPassword = '123456TEST';

    cy.visit('/register');
    fillForm(testEmail, testPassword, 'First Business', 'salon');
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/verify');
    enterVerificationCode('123456');
    cy.get('[data-cy="verify-submit"]').click();
    cy.url().should('include', '/login');

    cy.visit('/register');

    fillForm(testEmail, testPassword, 'Duplicate Business')
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/verify');

    enterVerificationCode('123456');
    cy.get('[data-cy="verify-submit"]').click();

    cy.get('[data-cy="error-message"]').should('be.visible');
    cy.get('[data-cy="error-message"]').should('not.be.empty');
    
    cy.task('db:countUsers').then((count) => {
      expect(count).to.equal(1);
    });
  });

  //test 3
  it('should show error for invalid verification code', () => {
    const testEmail = `invalid_${Date.now()}@example.com`;
    const testPassword = '123456TEST';
    const testBusinessName = `Test Business ${Date.now()}`;

    cy.visit('/register');
    fillForm(testEmail, testPassword, testBusinessName);
    cy.get('[data-cy="submit"]').click();
    cy.url().should('include', '/verify');

    enterVerificationCode('999999');
    cy.get('[data-cy="verify-submit"]').click();

    cy.get('[data-cy="error-message"]').should('be.visible');
    cy.get('[data-cy="error-message"]').should('not.be.empty');
    cy.url().should('include', '/verify');

    cy.task('db:findUserByEmail', testEmail).then((user) => {
      expect(user).to.be.null;
    });
  });
});

function fillForm(email, password, businessName, businessType = 'other') {
  cy.get('[data-cy="business-name"]').type(businessName);
  cy.get('[data-cy="email"]').type(email);
  cy.get('[data-cy="password"]').type(password);
  cy.get('[data-cy="business-type"]').select(businessType);
}

function enterVerificationCode(code) {
  for (let i = 0; i < code.length; i++) {
    cy.get(`[data-cy="verify-digit-${i}"]`).type(code[i]);
  }
}