const { defineConfig } = require("cypress");
const { Pool } = require("pg");

module.exports = defineConfig({
  e2e: {
    baseUrl: "http://localhost:5173",

    setupNodeEvents(on, config) {
      const pool = new Pool({
        host: String(config.env.DB_HOST || "localhost"),
        port: Number.parseInt(String(config.env.DB_PORT || "5432"), 10),
        database: String(config.env.DB_NAME || "online_queue_db"),
        user: String(config.env.DB_USER || "postgres"),
        password: String(config.env.DB_PASSWORD || "postgres"),
      });

      on("task", {
        "db:clean": async () => {
          await pool.query("DELETE FROM appointments");
          await pool.query("DELETE FROM clients");

          await pool.query(
            "DELETE FROM businesses WHERE NOT (name = $1 AND type = $2)",
            ["Demo Business", "service_company"]
          );

          await pool.query("DELETE FROM users WHERE login <> $1", ["admin@test.com"]);

          return null;
        },

        "db:cleanupDashboardTestClients": async () => {
          await pool.query(`
            DELETE FROM clients
            WHERE email LIKE 'dashboard_%@example.com'
               OR tg_username LIKE 'dashboard_%'
               OR phone LIKE '+79998%'
          `);

          return null;
        },

        "db:getDemoDashboardData": async () => {
          const result = await pool.query(`
            SELECT
              b.id AS business_id,
              br.id AS central_branch_id,
              br.name AS branch_name,
              br.address AS branch_address,
              br.timezone AS branch_timezone,
              e.id AS anna_employee_id,
              e.name AS employee_name,
              e.surname AS employee_surname,
              s.id AS consultation_service_id,
              s.name AS service_name,
              to_char(MIN(es.starts_at AT TIME ZONE br.timezone), 'YYYY-MM-DD') AS first_slot_date
            FROM businesses b
            JOIN branches br ON br.business_id = b.id
            JOIN employees e ON e.branch_id = br.id
            JOIN services s ON s.branch_id = br.id
            JOIN employee_services eps ON eps.employee_id = e.id AND eps.service_id = s.id
            JOIN employee_schedules es ON es.employee_id = e.id
            WHERE b.registration_slug = 'demo-business'
              AND br.name = 'Central Branch'
              AND e.name = 'Anna'
              AND e.surname = 'Petrova'
              AND s.name = 'Consultation'
              AND es.starts_at > now()
            GROUP BY b.id, br.id, e.id, s.id
            ORDER BY MIN(es.starts_at)
            LIMIT 1
          `);

          const row = result.rows[0] || null;

          if (!row) {
            return null;
          }

          return {
            business_id: Number(row.business_id),
            central_branch_id: Number(row.central_branch_id),
            branch_name: row.branch_name,
            branch_address: row.branch_address,
            branch_timezone: row.branch_timezone,
            anna_employee_id: Number(row.anna_employee_id),
            employee_name: row.employee_name,
            employee_surname: row.employee_surname,
            consultation_service_id: Number(row.consultation_service_id),
            service_name: row.service_name,
            first_slot_date: row.first_slot_date,
          };
        },

        "db:findAppointmentByPhone": async (phone) => {
          const result = await pool.query(
            `
              SELECT
                a.id,
                a.status,
                a.comment,
                c.phone AS client_phone,
                c.email AS client_email,
                c.name AS client_name,
                c.surname AS client_surname,
                br.name AS branch_name,
                s.name AS service_name,
                e.name AS employee_name,
                e.surname AS employee_surname
              FROM appointments a
              JOIN clients c ON c.id = a.client_id
              JOIN branches br ON br.id = a.branch_id
              JOIN services s ON s.id = a.service_id
              JOIN employees e ON e.id = a.employee_id
              WHERE c.phone = $1
              ORDER BY a.id DESC
              LIMIT 1
            `,
            [phone]
          );

          return result.rows[0] || null;
        },

        "db:findUserByEmail": async (email) => {
          const res = await pool.query("SELECT * FROM users WHERE login=$1", [email]);
          const user = res.rows[0] || null;

          if (user) {
            user.id = Number.parseInt(user.id, 10);
            user.business_id = Number.parseInt(user.business_id, 10);
          }

          return user;
        },

        "db:countUsers": async () => {
          const res = await pool.query(
            "SELECT COUNT(*) FROM users WHERE login <> $1",
            ["admin@test.com"]
          );

          return Number.parseInt(res.rows[0].count, 10);
        },

        "db:findBusinessByName": async (name) => {
          const result = await pool.query(
            "SELECT * FROM businesses WHERE name = $1",
            [name]
          );

          return result.rows[0] || null;
        },
      });

      return config;
    },
  },
});