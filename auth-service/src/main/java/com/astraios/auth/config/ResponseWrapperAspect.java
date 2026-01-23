package com.astraios.auth.config;

import com.astraios.auth.domain.vo.ApiResponse;
import lombok.extern.slf4j.Slf4j;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Pointcut;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import jakarta.servlet.http.HttpServletRequest;

/**
 * 使用 AOP 实现统一响应封装
 * 避免使用 ResponseBodyAdvice 的兼容性问题
 * 
 * 性能说明：
 * - AOP 与 ResponseBodyAdvice 的性能差异很小（通常 < 1ms）
 * - 对于 REST API，网络 I/O 是主要瓶颈，拦截器开销可忽略
 * - 如果追求极致性能，可以考虑使用 ResponseBodyAdvice（需要解决兼容性问题）
 */
@Slf4j
@Aspect
@Component
@Order(1)
public class ResponseWrapperAspect {

    /**
     * 定义切点：拦截 com.astraios.auth.controller 包下所有方法的返回值
     * 排除 JwksController 的 keys 方法
     */
    @Pointcut("execution(* com.astraios.auth.controller..*.*(..)) && " +
              "!execution(* com.astraios.auth.controller.JwksController.keys(..))")
    public void controllerMethods() {
    }

    /**
     * 环绕通知：在方法执行前后进行处理
     * 优化：先执行方法，再检查是否需要封装，减少路径检查的开销
     */
    @Around("controllerMethods()")
    public Object around(ProceedingJoinPoint joinPoint) throws Throwable {
        // 先执行原方法
        Object result = joinPoint.proceed();

        // 如果返回值已经是 ApiResponse，直接返回（避免重复封装）
        if (result instanceof ApiResponse) {
            return result;
        }

        // 检查是否是 Swagger 相关路径，如果是则不封装
        // 注意：这个检查放在方法执行后，因为大部分请求都是业务请求，不是 Swagger 请求
        HttpServletRequest request = getRequest();
        if (request != null) {
            String requestPath = request.getRequestURI();
            if (requestPath != null && isSwaggerPath(requestPath)) {
                // Swagger 相关路径，直接返回原始结果
                return result;
            }
        }

        // 封装成 ApiResponse
        return ApiResponse.success(result);
    }

    /**
     * 判断是否是 Swagger 相关路径
     * 提取为方法，便于优化和测试
     */
    private boolean isSwaggerPath(String requestPath) {
        return requestPath.startsWith("/v3/api-docs") ||
               requestPath.startsWith("/swagger-ui") ||
               requestPath.startsWith("/swagger-resources") ||
               requestPath.startsWith("/webjars") ||
               requestPath.contains("springdoc") ||
               requestPath.contains("openapi") ||
               requestPath.equals("/.well-known/jwks.json");
    }

    /**
     * 获取当前 HTTP 请求
     * 优化：使用局部变量缓存，减少重复获取
     */
    private HttpServletRequest getRequest() {
        try {
            ServletRequestAttributes attributes = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
            return attributes != null ? attributes.getRequest() : null;
        } catch (Exception e) {
            return null;
        }
    }
}

