-- Create the destinations table
CREATE TABLE IF NOT EXISTS destinations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

-- Create the bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    gender VARCHAR(10),
    birthday DATE,
    launchpad_id VARCHAR(50) NOT NULL,
    destination_id INTEGER REFERENCES destinations(id),
    launch_date DATE NOT NULL
);

-- Insert initial data into destinations
INSERT INTO destinations (name) VALUES
('Mars'),
('Moon'),
('Pluto'),
('Asteroid Belt'),
('Europa'),
('Titan'),
('Ganymede');