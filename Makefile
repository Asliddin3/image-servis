CURRENT_DIR=$(shell pwd)
APP=image
APP_CMD_DIR=./cmd

run:
	go run cmd/app/main.go
	
proto-gen:
	./scripts/gen-proto.sh	${CURRENT_DIR}

# migrate_up:
# 	migrate -source file://migrations -database postgres://abdulloh:abdulloh@database-1.c9lxq3r1itbt.us-east-1.rds.amazonaws.com:5432/customerdb_abdulloh?sslmode=disable up

# migrate_down:
# 	migrate -source file://migrations -database postgres://abdulloh:abdulloh@database-1.c9lxq3r1itbt.us-east-1.rds.amazonaws.com:5432/customerdb_abdulloh?sslmode=disable down

# migrate_force:
# 	migrate -path migrations/ -database postgres://abdulloh:abdulloh@database-1.c9lxq3r1itbt.us-east-1.rds.amazonaws.com:5432/customerdb_abdulloh?sslmode=disable force 1

migrate_images:
	migrate create -ext sql -dir migrations -seq create_images_table


migrate_up:
	migrate -path migrations/ -database postgres://postgres:compos1995@localhost:5432/imagedb?sslmode=disable up

migrate_down:
	migrate -path migrations/ -database postgres://postgres:compos1995@localhost:5432/imagedb?sslmode=disable down

migrate_force:
	migrate -path migrations/ -database postgres://postgres:compos1995@localhost:5432/imagedb?sslmode=disable force 1
