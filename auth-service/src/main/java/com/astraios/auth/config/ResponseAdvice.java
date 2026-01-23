package com.astraios.auth.config;

import com.astraios.auth.domain.vo.ApiResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.core.MethodParameter;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.http.converter.HttpMessageConverter;
import org.springframework.http.server.ServerHttpRequest;
import org.springframework.http.server.ServerHttpResponse;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;
import org.springframework.web.servlet.mvc.method.annotation.ResponseBodyAdvice;

import jakarta.servlet.http.HttpServletRequest;

/**
 * 统一响应封装拦截器
 * 自动将 controller 的返回值封装成统一的 ApiResponse 格式
 *
 */
@Slf4j
// @RestControllerAdvice  // 暂时注释掉，避免与 SpringDoc 冲突
public class ResponseAdvice implements ResponseBodyAdvice<Object> {
    
    /**
     * 获取当前 HTTP 请求
     */
    private HttpServletRequest getRequest() {
        try {
            ServletRequestAttributes attributes = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
            return attributes != null ? attributes.getRequest() : null;
        } catch (Exception e) {
            return null;
        }
    }

    /**
     * 判断是否需要对响应进行处理
     * @param returnType 返回类型
     * @param converterType 消息转换器类型
     * @return true 表示需要处理，false 表示不需要处理
     */
    @Override
    public boolean supports(MethodParameter returnType, Class<? extends HttpMessageConverter<?>> converterType) {
        return false;
    }

    /**
     * 在响应写入之前进行处理
     * @param body controller 返回的原始数据
     * @param returnType 返回类型
     * @param selectedContentType 选中的内容类型
     * @param selectedConverterType 选中的转换器类型
     * @param request 请求
     * @param response 响应
     * @return 处理后的响应体
     */
    @Override
    public Object beforeBodyWrite(Object body, MethodParameter returnType, MediaType selectedContentType,
                                  Class<? extends HttpMessageConverter<?>> selectedConverterType,
                                  ServerHttpRequest request, ServerHttpResponse response) {
        
        // 排除 Swagger 相关的路径，不进行封装
        String requestPath = request.getURI().getPath();
        if (requestPath != null) {
            // 检查是否是 Swagger/SpringDoc 相关的路径
            if (requestPath.startsWith("/v3/api-docs") ||
                requestPath.startsWith("/swagger-ui") ||
                requestPath.startsWith("/swagger-resources") ||
                requestPath.startsWith("/webjars") ||
                requestPath.equals("/.well-known/jwks.json") ||
                requestPath.contains("springdoc") ||
                requestPath.contains("openapi")) {
                log.debug("排除 Swagger 路径: {}", requestPath);
                return body;
            }
        }
        
        // 检查返回类型，如果是 SpringDoc 相关的类型，不处理
        if (returnType != null && returnType.getDeclaringClass() != null) {
            String className = returnType.getDeclaringClass().getName();
            if (className.contains("springdoc") || className.contains("org.springdoc")) {
                log.debug("排除 SpringDoc 类: {}", className);
                return body;
            }
        }
        
        // 如果返回值已经是 ApiResponse，直接返回（避免重复封装）
        if (body instanceof ApiResponse) {
            return body;
        }
        
        // 检查 body 是否是 OpenAPI 文档（Map 类型且包含 openapi 字段）
        if (body instanceof java.util.Map) {
            @SuppressWarnings("unchecked")
            java.util.Map<String, Object> bodyMap = (java.util.Map<String, Object>) body;
            if (bodyMap.containsKey("openapi") || bodyMap.containsKey("swagger")) {
                log.debug("检测到 OpenAPI 文档，不进行封装");
                return body;
            }
        }
        
        // 如果返回值是 ResponseEntity，提取 body
        if (body instanceof ResponseEntity) {
            ResponseEntity<?> responseEntity = (ResponseEntity<?>) body;
            Object responseBody = responseEntity.getBody();
            
            // 如果 ResponseEntity 的 body 已经是 ApiResponse，直接返回
            if (responseBody instanceof ApiResponse) {
                return responseBody;
            }
            
            // 检查 responseBody 是否是 OpenAPI 文档
            if (responseBody instanceof java.util.Map) {
                @SuppressWarnings("unchecked")
                java.util.Map<String, Object> bodyMap = (java.util.Map<String, Object>) responseBody;
                if (bodyMap.containsKey("openapi") || bodyMap.containsKey("swagger")) {
                    log.debug("检测到 OpenAPI 文档（ResponseEntity），不进行封装");
                    return responseBody;
                }
            }
            
            // 封装成 ApiResponse
            return ApiResponse.success(responseBody);
        }
        
        // 如果返回值为 null，也封装成 ApiResponse
        if (body == null) {
            return ApiResponse.success(null);
        }
        
        // 其他情况，直接封装成 ApiResponse
        return ApiResponse.success(body);
    }
}

