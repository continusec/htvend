FROM docker.io/library/maven:3.9.11-amazoncorretto-21-debian AS build

ADD https://api.github.com/repos/spring-guides/gs-spring-boot/tarball/main /src.tar.gz

RUN mkdir -p /src && \
    tar -C /src --strip-components=1 -zxf /src.tar.gz && \
    cd /src/complete && \
    mvn package -Dproject.build.outputTimestamp=2023-01-01T00:00:00Z

FROM docker.io/library/amazoncorretto:21.0.8

EXPOSE 8080

COPY --from=build /src/complete/target/*.jar /app.jar

ENTRYPOINT ["java", "-jar", "/app.jar"]
