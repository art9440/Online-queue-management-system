describe('Registration Page: e2e testing ', () => {
  beforeEach(() => {
    cy.visit('/register');
  });

  it('should successfully register and redirect to verify', ()=> {
    const testEmail = `test_${Date.now()}@example.com`;
    fillForm(testEmail);
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/verify');
    cy.window().its('localStorage.REGISTRATION_ID').should('exist');
  });

  it('should show error when email is already exists', () => {
    const testEmail = `test_${Date.now()}@example.com`;
    fillForm(testEmail);
    cy.get('[data-cy="submit"]').click();

    cy.url().should('include', '/verify');
    cy.visit('/register');
    fillForm(testEmail);
    cy.get('[data-cy="submit"]').click();

    cy.url().should('not.include', '/verify')
  });

  it('should show validation error with empty form', () => {
    cy.get('[data-cy="submit"]').click();
    cy.contains('Введите название бизнеса').should('be.visible');
    cy.contains('Введите email').should('be.visible');
    cy.contains('Введите пароль').should('be.visible');

    cy.url().should('not.include', '/verify');
    cy.url().should('include', '/register');
  });

  it('should show email error for invalid format', () => {
    fillForm('invalid-email');
    cy.get('[data-cy="submit"]').click();

    cy.contains('Введите корректный email').should('be.visible');
    cy.url().should('not.include', '/verify');
  });

  it('should show password error for too short password', () => {
    cy.get('[data-cy="business-name"]').type('Test Business');
    cy.get('[data-cy="email"]').type('test@example.com');
    cy.get('[data-cy="password"]').type('123');
    cy.get('[data-cy="submit"]').click();
    
    cy.contains('Минимум 6 символов').should('be.visible');
    cy.url().should('not.include', '/verify');
  });

  it('shouls show error for short business name', () => {
    cy.get('[data-cy="business-name"]').type('A');
    cy.get('[data-cy="email"]').type('test@example.com');
    cy.get('[data-cy="password"]').type('password123');
    cy.get('[data-cy="submit"]').click();
    
    cy.contains('Минимум 2 символа').should('be.visible');
    cy.url().should('not.include', '/verify');
  });

  it('should show errors on blur', () => {
    cy.get('[data-cy="business-name"]').focus().blur();
    cy.get('[data-cy="email"]').focus().blur();
    cy.get('[data-cy="password"]').focus().blur();
    
    cy.contains('Введите название бизнеса').should('be.visible');
    cy.contains('Введите email').should('be.visible');
    cy.contains('Введите пароль').should('be.visible');
  });

  it('should clear errors when user starts typing', () => {
    cy.get('[data-cy="submit"]').click();
    cy.contains('Введите email').should('be.visible');
    
    cy.get('[data-cy="email"]').type('test@example.com');
    cy.contains('Введите email').should('not.exist');
  });

  it('should show network error with disconnect', () => {
    const testEmail = `test_${Date.now()}@example.com`;
    fillForm(testEmail);
    cy.intercept('POST', '/api/register', { forceNetworkError: true });
    cy.get('[data-cy="submit"]').click();
    cy.contains('Сервер не отвечает. Проверьте соединение.').should('be.visible');
    cy.url().should('not.include', '/verify');
    cy.url().should('include', '/register');
  });

  it('should handle 500 server error', () => {
    const testEmail = `test_${Date.now()}@example.com`;

    cy.intercept('POST', '/api/register', {
    statusCode: 500,
    body: { error: 'Internal Server Error' }
    });

    fillForm(testEmail);
    cy.get('[data-cy="submit"]').click();

    cy.contains('Internal Server Error').should('be.visible');
    cy.url().should('not.include', '/verify');
  });

  it('should not send duplicate requests when button is clicked multiple times', () => {
  const testEmail = `test_${Date.now()}@example.com`;
  
  let requestCount = 0;
  cy.intercept('POST', '/api/register', (req) => {
    requestCount++;
    req.reply({
      statusCode: 201,
      body: { registration_id: 'test-id-123' },
      delay: 2000
    });
  }).as('registerRequest');
  
  fillForm(testEmail);
  cy.get('[data-cy="submit"]').click();
  cy.get('[data-cy="submit"]').should('be.disabled');
  
  cy.wait('@registerRequest');
  cy.then(() => {
    expect(requestCount).to.equal(1);
  });
  
  cy.get('[data-cy="submit"]').should('be.disabled');
  cy.url().should('include', '/verify');
});

});

function fillForm(testExample) {
  cy.get('[data-cy="business-name"]').type('Test BusinessName');
  cy.get('[data-cy="email"]').type(testExample);
  cy.get('[data-cy="password"]').type('password123');
  cy.get('[data-cy="business-type"]').select('salon');
}