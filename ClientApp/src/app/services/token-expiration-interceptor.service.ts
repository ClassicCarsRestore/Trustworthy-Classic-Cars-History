import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { Router } from '@angular/router';
import { AuthService } from './auth.service';

@Injectable()
export class TokenExpirationInterceptor implements HttpInterceptor {
  constructor(private router: Router, private authService: AuthService) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(request).pipe(
      tap(
        (event) => {},
        (error) => {
          if (error && error.error && error.error.message === 'Failed to authenticate token. Make sure to include the token returned from /users call in the authorization header  as a Bearer token') {
            alert('Your session has expired. Please log in again.');
            this.authService.logout();
          }
        }
      )
    );
  }
}
