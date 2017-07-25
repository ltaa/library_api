CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS authors (
	author_id SERIAL PRIMARY KEY,
	first_name varchar(50),
	last_name varchar(50)
	-- book_id references authors_books(book_id)?
);


CREATE TABLE IF NOT EXISTS publishers (
	publisher_id SERIAL PRIMARY KEY,
	publisher_name varchar(50)
);


CREATE TABLE IF NOT EXISTS books (
	book_id SERIAL PRIMARY KEY,
	book_name varchar(50),
	year INTEGER,
	-- author_id SERIAL references authors_books(author_id),?
	publisher_id SERIAL references publishers(publisher_id)
);


CREATE TABLE IF NOT EXISTS authors_books (
	author_id SERIAL references authors(author_id),
	book_id SERIAL references books(book_id)
);

CREATE TYPE book_state AS ENUM ('library', 'processing', 'client');

CREATE TABLE IF NOT EXISTS book_instances (
	instance_id SERIAL PRIMARY KEY,
	book_id SERIAL references books (book_id),
	state book_state default 'library'
);


CREATE TABLE IF NOT EXISTS workers (
	worker_id SERIAL PRIMARY KEY,
	first_name varchar(50),
	last_name varchar(50),
	login varchar(50),
	password varchar(128)
);

CREATE TABLE IF NOT EXISTS operations (
	instance_id SERIAL references book_instances(instance_id),
	worker_id SERIAL references workers(worker_id),
	duration TIMESTAMP
);


CREATE TABLE IF NOT EXISTS clients (
	client_id SERIAL PRIMARY KEY,
	first_name varchar(50),
	last_name varchar(50)

);


CREATE TABLE IF NOT EXISTS readers (
	client_id SERIAL references clients(client_id),
	instance_id SERIAL references book_instances(instance_id),
	date_issue date,
	return_date date
);



-- CREATE FUNCTION update_book_state() RETURNS TRIGGER AS $update_book_state$
--   BEGIN
--     IF OLD.state = 'library' AND NEW.state != 'processing' THEN
--       RAISE EXCEPTION 'invalid transition';
-- 		END IF;
--     IF OLD.state = 'processing' AND NEW.state != 'client' THEN
--       RAISE EXCEPTION 'invalid transition';
-- 		END IF;
--     IF OLD.state = 'client' AND NEW.state != 'library' THEN
--       RAISE EXCEPTION 'invalid transition';
--     END IF;
-- 		RETURN NEW;
-- 	END;
-- $update_book_state$ LANGUAGE plpgsql;


-- CREATE TRIGGER update_book_state BEFORE UPDATE ON book_instances
--     FOR EACH ROW EXECUTE PROCEDURE update_book_state();




insert into authors (first_name, last_name ) VALUES ('Dennis', 'Ritchie');
insert into authors (first_name, last_name ) VALUES ('Brian', 'Kernighan');

insert into authors (first_name, last_name ) VALUES ('Андрей', 'Александреску');
insert into authors (first_name, last_name ) VALUES ('Ричард', 'Докинз');



insert into publishers (publisher_name ) VALUES ('dmk press');
insert into publishers (publisher_name ) VALUES ('аст');
insert into publishers (publisher_name ) VALUES ('williams');


insert into books(book_name,year,publisher_id) VALUES ('The C Programming Language', 2017, (select publisher_id from publishers where publisher_name='dmk press'));
insert into books(book_name,year,publisher_id) VALUES ('Современное проектирование на C++', 2015, (select publisher_id from publishers where publisher_name='williams'));
insert into books(book_name,year,publisher_id) VALUES ('бог как илюзия', 2016, (select publisher_id from publishers where publisher_name='аст'));

insert into book_instances (book_id) VALUES('1');
insert into book_instances (book_id) VALUES('1');
insert into book_instances (book_id) VALUES('1');
insert into book_instances (book_id) VALUES('2');
insert into book_instances (book_id) VALUES('3');
insert into book_instances (book_id) VALUES('3');



insert into clients (first_name,last_name) VALUES ('vasily', 'pupkin');
insert into clients (first_name,last_name) VALUES ('petr', 'petrov');
insert into clients (first_name,last_name) VALUES ('sidor', 'sidorov');

insert into workers(first_name, last_name, login, password) VALUES ('admin', 'admin', 'admin', crypt('admin', gen_salt('bf')));
insert into workers(first_name, last_name, login, password) VALUES ('gena', 'bukin', 'test', crypt('test', gen_salt('bf')));


insert into authors_books(author_id,book_id) VALUES ((select author_id from authors where last_name = 'Ritchie'), (select book_id from books where book_name = 'The C Programming Language'));
insert into authors_books(author_id,book_id) VALUES ((select author_id from authors where last_name = 'Kernighan'), (select book_id from books where book_name = 'The C Programming Language'));
insert into authors_books(author_id,book_id) VALUES ((select author_id from authors where last_name = 'Александреску'), (select book_id from books where book_name = 'Современное проектирование на C++'));
insert into authors_books(author_id,book_id) VALUES ((select author_id from authors where last_name = 'Докинз'), (select book_id from books where book_name = 'бог как илюзия'));
