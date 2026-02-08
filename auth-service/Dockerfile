# Build stage (non-Alpine to ensure protoc-gen-grpc-java runs with glibc)
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app

# Copy the parent pom
COPY pom.xml .

# Copy shared proto definitions (required for gRPC codegen)
COPY proto/ proto/

# Copy the specific service directory
COPY auth-service/ auth-service/

# If there are other shared modules like common-service that auth-service depends on, 
# they should be copied here too. For now, we assume only parent and the service itself.

# Build the specific service
# -pl auth-service: build only this project
# -am: also make dependencies (if any internal modules are needed)
RUN mvn clean package -pl auth-service -am -DskipTests

# Run stage
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=build /app/auth-service/target/auth-*.jar app.jar
EXPOSE 9091
ENTRYPOINT ["java", "-jar", "app.jar"]
