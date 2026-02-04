package com.astraios.auth.config;

import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import lombok.extern.slf4j.Slf4j;
import org.springdoc.core.customizers.GlobalOpenApiCustomizer;
import org.springdoc.core.models.GroupedOpenApi;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpHeaders;


/**
 * Swagger配置类
 * 注意：由于项目已改为gRPC架构，不再需要REST API文档，此类已禁用
 */
// @Configuration
@Slf4j
public class SwaggerConfig {
    @Value(value = "${host:localhost}")
    private String host;

    @Value("${server.port:8081}")
    private String port;
    
    /**
     * 配置 API 分组，只扫描当前项目的 controller
     */
    @Bean
    public GroupedOpenApi publicApi() {
        return GroupedOpenApi.builder()
                .group("astraios-api")
                .pathsToMatch("/api/**")
                .packagesToScan("com.astraios.auth.controller")
                .build();
    }

    @Bean
    public GlobalOpenApiCustomizer orderGlobalOpenApiCustomizer() {
        return openApi -> {
            // 全局添加鉴权参数
            if(openApi.getPaths()!=null){
                openApi.getPaths().forEach((s, pathItem) -> {
                    // 为所有接口添加鉴权
                    pathItem.readOperations().forEach(operation -> {
                        operation.addSecurityItem(new SecurityRequirement().addList(HttpHeaders.AUTHORIZATION));
                    });
                });
            }

        };
    }

    @Bean
    public OpenAPI springShopOpenAPI() {
        log.info("http://{}:{}/swagger-ui.html", host, port);

        return new OpenAPI()
                .info(new Info()
                        .title("astraios API")
                        .description("astraios API文档")
                        .version("v1")
                        .license(new License().name("Apache 2.0").url("http://springdoc.org")))
                .addSecurityItem(new SecurityRequirement().addList(HttpHeaders.AUTHORIZATION))
                .components(new Components().addSecuritySchemes(HttpHeaders.AUTHORIZATION,new SecurityScheme()
                        .name(HttpHeaders.AUTHORIZATION).type(SecurityScheme.Type.HTTP).scheme("bearer")));
    }

}