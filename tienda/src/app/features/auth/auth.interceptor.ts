import { Injectable } from '@angular/core'; import { HttpInterceptor, HttpRequest, HttpHandler } from '@angular/common/http';
@Injectable() export class AuthInterceptor implements HttpInterceptor { intercept(request: HttpRequest<unknown>, next: HttpHandler) { return next.handle(request.clone({ withCredentials: true })); } }
