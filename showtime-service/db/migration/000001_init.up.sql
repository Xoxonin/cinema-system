CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    capacity INT NOT NULL
);

CREATE TABLE showtimes (
    id SERIAL PRIMARY KEY,
    movie_id INT NOT NULL,
    room_id INT NOT NULL REFERENCES rooms(id),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL
);
