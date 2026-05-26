
describe('E2E tests: login flow', () => {
  beforeEach(() => {
    cy.task('db:clean');
  });

  //test 1
  it('full path: register, verify, login, dashboard', () => {
    const unique = Date.now();
    const testEmail = `reg_${unique}@example.com`;
    const testSecret = createTestSecret(unique);
    const testBusinessName = `Test Business ${unique}`;

    cy.visit('/register');
    fillForm(testEmail, testSecret, testBusinessName);
    cy.get('[data-cy="submit"]').click();
    cy.url().should('include', '/verify');
    enterVerificationCode("123456");
    cy.get('[data-cy="verify-submit"]').click();

    cy.url().should('include', '/login');
    cy.get('[data-cy="login"]').type(testEmail);
    cy.get('[data-cy="password"]').type(testSecret);
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/admin');
    cy.task('db:findUserByEmail', testEmail).then(user => {
      expect(user).to.not.be.null;
      expect(user.login).to.eq(testEmail);
    });
  });

  //test 2
  it('should show error for wrong password', () => {
    const unique = Date.now();
    const testEmail = `wrong_pass_${unique}@example.com`;
    const correctSecret = createTestSecret(unique);
    const wrongSecret = createTestSecret(`${unique}_wrong`);
    const testBusinessName = `Wrong Password Business ${unique}`;
    
    cy.visit('/register');
    fillForm(testEmail, correctSecret, testBusinessName, 'salon');
    cy.get('[data-cy="submit"]').click();
    cy.url().should('include', '/verify');

    enterVerificationCode("123456");
    cy.get('[data-cy="verify-submit"]').click();
    cy.url().should('include', '/login');

    cy.get('[data-cy="login"]').type(testEmail);
    cy.get('[data-cy="password"]').type(wrongSecret);
    cy.get('[data-cy="submit"]').click();

    cy.get('[data-cy="error-message"]', { timeout: 5000 }).should('be.visible');

    cy.get('[data-cy="login"]').should('have.value', testEmail);
    cy.get('[data-cy="password"]').should('have.value', wrongSecret);
  });

  //test 3
  it('should handle non-existent login', () => {
    const unique = Date.now();
    const testEmail = `nonexistent_${unique}@example.com`;
    const testSecret = createTestSecret(`${unique}_missing`);

    cy.visit('/login');
    cy.get('[data-cy="login"]').type(testEmail);
    cy.get('[data-cy="password"]').type(testSecret);
    cy.get('[data-cy="submit"]').click();

    cy.get('[data-cy="error-message"]').should('be.visible');
  });
});

function fillForm(email, password, businessName, businessType = 'other') {
  cy.get('[data-cy="business-name"]').type(businessName);
  cy.get('[data-cy="email"]').type(email);
  cy.get('[data-cy="password"]').type(password);
  cy.get('[data-cy="business-type"]').select(businessType);
}

function enterVerificationCode(code) {
  for (let i = 0; i < code.length; i += 1) {
    cy.get(`[data-cy="verify-digit-${i}"]`).type(code[i]);
  }
}

function createTestSecret(seed) {
  return ["E2e", seed, "Qw9!"].join("-");
}
