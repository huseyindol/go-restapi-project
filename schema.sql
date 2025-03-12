CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Örnek veri
INSERT INTO users (name, email) VALUES 
  ('Hüseyin DOL', 'info@huseyindol.site'),
  ('Yağız Efe DOL', 'efe@huseyindol.site');

