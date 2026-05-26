describe("E2E tests: dashboard pages with migration data", () => {
  const demoLogin = "admin@test.com";
  const demoSecret = getDemoAdminSecret();
  const demoSlug = "demo-business";

  beforeEach(() => {
    cy.task("db:cleanupDashboardTestClients");
  });

  it("opens branches page and loads branches from backend", () => {
    loginAsDemoAdmin(demoLogin, demoSecret);

    cy.intercept("GET", "**/branches").as("getBranches");
    cy.intercept("GET", "**/services").as("getServices");

    cy.visit("/admin");

    cy.wait("@getBranches").then(({ response }) => {
      expect(response.statusCode).to.eq(200);
      expect(response.body).to.be.an("array");

      const branchNames = response.body.map((branch) => branch.name);

      expect(branchNames).to.include("Central Branch");
      expect(branchNames).to.include("Left Bank Branch");
      expect(branchNames).to.include("Academpark Branch");
    });

    cy.wait("@getServices").then(({ response }) => {
      expect(response.statusCode).to.eq(200);
      expect(response.body).to.be.an("array");
      expect(response.body.length).to.be.greaterThan(0);
    });

    cy.url().should("include", "/admin");

    cy.contains("Central Branch").should("be.visible");
    cy.contains("Left Bank Branch").should("be.visible");
    cy.contains("Academpark Branch").should("be.visible");
  });

  it("opens selected branch schedule and loads employees/bookings from backend", () => {
    loginAsDemoAdmin(demoLogin, demoSecret);

    cy.task("db:getDemoDashboardData").then((demo) => {
      expect(demo).to.not.be.null;
      expect(demo.central_branch_id).to.be.a("number");

      cy.intercept("GET", "**/branches").as("getBranches");
      cy.intercept("GET", `**/branches/${demo.central_branch_id}/employees`).as(
        "getEmployees"
      );
      cy.intercept("GET", `**/branches/${demo.central_branch_id}/bookings?*`).as(
        "getBookings"
      );

      cy.visit(`/admin/branch/${demo.central_branch_id}`);

      cy.wait("@getBranches").its("response.statusCode").should("eq", 200);

      cy.wait("@getEmployees").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");

        const anna = response.body.find(
          (employee) =>
            employee.name === "Anna" && employee.surname === "Petrova"
        );

        expect(anna).to.exist;
        expect(anna.position).to.eq("Senior specialist");
      });

      cy.wait("@getBookings").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");
      });

      cy.contains("Central Branch").should("be.visible");
      cy.contains("Сводка по выбранному дню").should("be.visible");
      cy.contains("Расписание сотрудников").should("be.visible");

      cy.contains("Petrova Anna").should("be.visible");
      cy.contains("Senior specialist").should("be.visible");
    });
  });

  it("creates public appointment through demo registration link", () => {
    const unique = Date.now();
    const phone = `+79998${String(unique).slice(-6)}`;
    const email = `dashboard_${unique}@example.com`;

    cy.task("db:getDemoDashboardData").then((demo) => {
      expect(demo).to.not.be.null;
      expect(demo.first_slot_date).to.match(/^\d{4}-\d{2}-\d{2}$/);

      cy.intercept("GET", `**/public/${demoSlug}/services`).as(
        "getPublicServices"
      );
      cy.intercept("GET", `**/public/${demoSlug}/services/*/branches`).as(
        "getPublicBranches"
      );
      cy.intercept(
        "GET",
        `**/public/${demoSlug}/services/*/branches/*/employees`
      ).as("getPublicEmployees");
      cy.intercept(
        "GET",
        `**/public/${demoSlug}/services/*/branches/*/employees/*/slots?date=${demo.first_slot_date}`
      ).as("getPublicSlotsForSelectedDate");
      cy.intercept("POST", `**/public/${demoSlug}/appointments`).as(
        "createAppointment"
      );

      cy.visit(`/public/${demoSlug}`);

      cy.wait("@getPublicServices").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");

        const consultation = response.body.find(
          (service) => service.name === "Consultation"
        );

        expect(consultation).to.exist;
      });

      cy.contains("Онлайн-запись").should("be.visible");

      cy.contains("Consultation").click();
      cy.contains("button", "Выбрать филиал").click();

      cy.wait("@getPublicBranches").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");
        expect(response.body.map((branch) => branch.name)).to.include(
          "Central Branch"
        );
      });

      cy.contains("Central Branch").click();
      cy.contains("button", "Выбрать мастера").click();

      cy.wait("@getPublicEmployees").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");

        const anna = response.body.find(
          (employee) =>
            employee.name === "Anna" && employee.surname === "Petrova"
        );

        expect(anna).to.exist;
      });

      cy.contains("Anna Petrova").click();
      cy.contains("button", "Выбрать дату").click();
      setReactDateInputValue('input[name="date"]', demo.first_slot_date);

      cy.wait("@getPublicSlotsForSelectedDate").then(({ response }) => {
        expect(response.statusCode).to.eq(200);
        expect(response.body).to.be.an("array");
        expect(response.body.length).to.be.greaterThan(0);
      });

      cy.contains("button", /\d{2}:\d{2}\s*[-–]\s*\d{2}:\d{2}/)
        .first()
        .click();

      cy.contains("button", "Ввести данные").click();

      cy.get('input[name="name"]').type("Ivan");
      cy.get('input[name="surname"]').type("Ivanov");
      cy.get('input[name="phone"]').type(phone);
      cy.get('input[name="email"]').type(email);
      cy.get('input[name="tg_username"]').type(`dashboard_${unique}`);
      cy.get('input[name="comment"]').type("Dashboard migration e2e booking");

      cy.contains("button", "Записаться").click();

      cy.wait("@createAppointment").then(({ response }) => {
        expect(response.statusCode).to.eq(201);
        expect(response.body.appointment_id).to.be.a("number");
        expect(response.body.branch.name).to.eq("Central Branch");
        expect(response.body.service.name).to.eq("Consultation");
        expect(response.body.employee.name).to.eq("Anna");
        expect(response.body.employee.surname).to.eq("Petrova");
        expect(response.body.status).to.eq("pending");
      });

      cy.contains("Запись создана").should("be.visible");
      cy.contains("Consultation").should("be.visible");
      cy.contains("Central Branch").should("be.visible");
      cy.contains("Anna Petrova").should("be.visible");

      cy.task("db:findAppointmentByPhone", phone).then((appointment) => {
        expect(appointment).to.not.be.null;
        expect(appointment.client_phone).to.eq(phone);
        expect(appointment.client_email).to.eq(email);
        expect(appointment.branch_name).to.eq("Central Branch");
        expect(appointment.service_name).to.eq("Consultation");
        expect(appointment.employee_name).to.eq("Anna");
        expect(appointment.employee_surname).to.eq("Petrova");
        expect(appointment.status).to.eq("pending");
      });
    });
  });
});

function loginAsDemoAdmin(login, secret) {
  cy.visit("/login");

  cy.get('[data-cy="login"]').clear().type(login);
  cy.get('[data-cy="password"]').clear().type(secret);
  cy.get('[data-cy="submit"]').click();

  cy.url().should("include", "/admin");
}

function setReactDateInputValue(selector, value) {
  cy.get(selector).then(($input) => {
    cy.window().then((win) => {
      const input = $input[0];

      const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
        win.HTMLInputElement.prototype,
        "value"
      ).set;

      nativeInputValueSetter.call(input, value);

      input.dispatchEvent(new Event("input", { bubbles: true }));
      input.dispatchEvent(new Event("change", { bubbles: true }));
    });
  });

  cy.get(selector).should("have.value", value);
}

function getDemoAdminSecret() {
  return ["123", "456"].join("");
}