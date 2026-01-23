package com.astraios.auth.handler;

import com.astraios.auth.domain.vo.ApiResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import io.swagger.v3.oas.annotations.Hidden;

/**
 * 全局异常处理器
 * 统一处理异常并返回统一的响应格式
 */
@Slf4j
@RestControllerAdvice
@Hidden
public class GlobalExceptionHandler {

    /**
     * 处理运行时异常
     */
    @ExceptionHandler(RuntimeException.class)
    @ResponseStatus(HttpStatus.OK)
    public ApiResponse<?> handleRuntimeException(RuntimeException e) {
        log.error("RuntimeException: ", e);
        return ApiResponse.fail(-1, e.getMessage());
    }

    /**
     * 处理认证异常
     */
    @ExceptionHandler(BadCredentialsException.class)
    @ResponseStatus(HttpStatus.OK)
    public ApiResponse<?> handleBadCredentialsException(BadCredentialsException e) {
        log.error("BadCredentialsException: ", e);
        return ApiResponse.fail(-1, "Invalid username or password");
    }

    /**
     * 处理所有其他异常
     */
    @ExceptionHandler(Exception.class)
    @ResponseStatus(HttpStatus.OK)
    public ApiResponse<?> handleException(Exception e) {
        log.error("Exception: ", e);
        return ApiResponse.fail(-1, "Internal server error");
    }
}

