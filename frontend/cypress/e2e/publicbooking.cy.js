describe('Public Booking Page: e2e testing', () => {
    beforeEach(() => {
        cy.intercept(
            'GET',
            '**/services**'
        ).as('getServices');

        cy.intercept(
            'GET',
            '/public/**/services/**/branches**'
        ).as('getBranches');

        cy.intercept(
            'GET',
            '/public/**/services/**/branches/**/employees'
        ).as('getEmployees');

        cy.intercept(
            'GET',
            '**/slots**'
        ).as('getSlots');

        cy.intercept(
            'POST',
            '**/appointments'
        ).as('createAppointment');

        cy.visit('/public/demo-business');
    });

    it('should successfully create a booking', () => {
        completeStepsToDataForm();

        // Form
        fillDataForm();
        cy.get('[data-cy="confirm-booking-button"]').click();

        // Confirm appointment creation
        cy.wait('@createAppointment')
        .its('response.statusCode')
        .should('eq', 201);

        cy.get('[data-cy="google-calendar-button"]')
        .should('exist');
    });

    it('should validate required fields in data form', () => {
        completeStepsToDataForm();
        
        cy.get('[data-cy="confirm-booking-button"]').click();

        cy.contains('Введите имя').should('exist');
        cy.contains('Введите фамилию').should('exist');
        cy.contains('Введите телефон').should('exist');
    });

    it('should validate email format', () => {
        completeStepsToDataForm();
        
        cy.get('[data-cy="data-name-input"]').type('Ivan');
        cy.get('[data-cy="data-surname-input"]').type('Ivanov');
        cy.get('[data-cy="data-phone-input"]').type('+78005553535');
        cy.get('[data-cy="data-email-input"]').type('invalid-email');
        
        cy.get('[data-cy="confirm-booking-button"]').click();
        
        cy.contains('Неверный формат email').should('exist');
    });

    it('should validate phone format', () => {
        completeStepsToDataForm();
        
        cy.get('[data-cy="data-phone-input"]').type('invalid-phone');
        cy.get('[data-cy="confirm-booking-button"]').click();
        
        // Проверяем ошибку валидации телефона
        cy.contains('Неверный формат телефона').should('exist');
    });

    it('should show there are no free slots', () => {
        cy.wait('@getServices');

        cy.contains('Загружаем услуги...').should('not.exist');

        cy.get('[data-cy="service-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-branch-button"]').click();

        cy.wait('@getBranches');

        cy.get('[data-cy="branch-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-master-button"]').click();

        cy.wait('@getEmployees');

        cy.get('[data-cy="master-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-time-button"]').click();

        const targetDate = new Date();

        targetDate.setDate(targetDate.getDate() + 8);

        const formattedDate =
            targetDate.toISOString().split('T')[0];

        cy.get('[data-cy="date-input"]')
        .invoke('val', formattedDate)
        .trigger('input')
        .trigger('change');

        cy.contains(
        'На выбранную дату свободных слотов нет'
        ).should('exist');

        cy.get('[data-cy="time-select-button"]')
        .should('not.exist');
        });


    it('should save previously saved data', () => {
        cy.wait('@getServices');

        cy.contains('Загружаем услуги...').should('not.exist');

        cy.get('[data-cy="service-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-branch-button"]').click();

        // Back button pressed
        cy.get('[data-cy="back-button"]').click();
        cy.get('[data-cy="service-option-button"]')
        .first()
        .should('have.class', 'border-blue-500')
        .and('have.class', 'bg-blue-50');

        cy.get('[data-cy="go-to-branch-button"]').click();

        cy.wait('@getBranches');

        cy.get('[data-cy="branch-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-master-button"]').click();

        // Back button pressed
        cy.get('[data-cy="back-button"]').click();
        cy.get('[data-cy="branch-option-button"')
        .first()
        .should('have.class', 'border-blue-500')
        .and('have.class', 'bg-blue-50');

        cy.get('[data-cy="go-to-master-button"]').click();

        cy.wait('@getEmployees');

        cy.get('[data-cy="master-option-button"]')
        .should('have.length.at.least', 1)
        .first()
        .click();
        cy.get('[data-cy="go-to-time-button"]').click();

        // Back button pressed
        cy.get('[data-cy="back-button"]').click();
        cy.get('[data-cy="master-option-button"]')
        .first()
        .should('have.class', 'border-blue-500')
        .and('have.class', 'bg-blue-50');

        cy.get('[data-cy="go-to-time-button"]').click();

        cy.wait('@getSlots');

        const targetDate = new Date();

        targetDate.setDate(targetDate.getDate() + 2);

        const formattedDate =
            targetDate.toISOString().split('T')[0];

        cy.get('[data-cy="date-input"]')
    .then(($input) => {
        const input = $input[0];

        const nativeInputValueSetter =
            Object.getOwnPropertyDescriptor(
                window.HTMLInputElement.prototype,
                'value'
            ).set;

        nativeInputValueSetter.call(
            input,
            formattedDate
        );

        input.dispatchEvent(
            new Event('input', { bubbles: true })
        );

        input.dispatchEvent(
            new Event('change', { bubbles: true })
        );
    });

        cy.get('[data-cy="date-input"]')
        .should('have.value', formattedDate);

        cy.get('[data-cy="time-select-button"]')
        .should('have.length.greaterThan', 0)
        .first()
        .click();

        cy.get('[data-cy="go-to-data-button"]')
        .click();

        // Back button pressed
        cy.get('[data-cy="back-button"]').click();
        cy.get('[data-cy="time-select-button"]')
        .first()
        .should('have.class', 'border-blue-500')
        .and('have.class', 'bg-blue-50');

        cy.get('[data-cy="go-to-data-button"]')
        .click();

        fillDataForm();
        cy.get('[data-cy="confirm-booking-button"]').click();

        cy.wait('@createAppointment')
        .its('response.statusCode')
        .should('eq', 201);
    });

});

function completeStepsToDataForm() {
    cy.wait('@getServices');
    // Services
    cy.contains('Загружаем услуги...').should('not.exist');

    cy.get('[data-cy="service-option-button"]')
    .should('have.length.at.least', 1)
    .first()
    .click();
    cy.get('[data-cy="go-to-branch-button"]').click();

    // Branches
    cy.wait('@getBranches');

    cy.get('[data-cy="branch-option-button"]')
    .should('have.length.at.least', 1)
    .first()
    .click();
    cy.get('[data-cy="go-to-master-button"]').click();

    // Employees
    cy.wait('@getEmployees');

    cy.get('[data-cy="master-option-button"]')
    .should('have.length.at.least', 1)
    .first()
    .click();
    cy.get('[data-cy="go-to-time-button"]').click();

    // Slots (time & data)
    cy.wait('@getSlots');

    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 2);
    const formattedDate = tomorrow.toISOString().split('T')[0];

    cy.get('[data-cy="date-input"]')
    .then(($input) => {
        const input = $input[0];

        const nativeInputValueSetter =
            Object.getOwnPropertyDescriptor(
                window.HTMLInputElement.prototype,
                'value'
            ).set;

        nativeInputValueSetter.call(
            input,
            formattedDate
        );

        input.dispatchEvent(
            new Event('input', { bubbles: true })
        );

        input.dispatchEvent(
            new Event('change', { bubbles: true })
        );
    });

    cy.get('[data-cy="date-input"]')
    .should('have.value', formattedDate);

    cy.get('[data-cy="time-select-button"]')
    .should('have.length.greaterThan', 0)
    .first()
    .click();

    cy.get('[data-cy="go-to-data-button"]')
    .click();
}

function fillDataForm() {
    cy.get('[data-cy="data-name-input"]').type('Ivan');
    cy.get('[data-cy="data-surname-input"]').type('Ivanov');
    cy.get('[data-cy="data-phone-input"]').type('+78005553535');
    cy.get('[data-cy="data-email-input"]').type('test@example.com');
    cy.get('[data-cy="data-telegram-input"]').type('test12345');
    cy.get('[data-cy="data-comment-input"]').type('Мешать, но не взбалтывать');
}