import { Component } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { Observable } from 'rxjs';


@Component({
  selector: 'app-nav-menu',
  templateUrl: './nav-menu.component.html',
  styleUrls: ['./nav-menu.component.css']
})
export class NavMenuComponent {
  isExpanded = false;
  loggedInEmail$: Observable<string>;

  constructor(public authService: AuthService) {
    this.loggedInEmail$ = this.authService.loggedInEmailObservable;
  }

  ngOnInit() {
    this.loggedInEmail$ = this.authService.loggedInEmailObservable;
  }

  logout() {
    this.authService.logout();
  }

  collapse() {
    this.isExpanded = false;
  }

  toggle() {
    this.isExpanded = !this.isExpanded;
  }

  isLoggedIn(): boolean {
    return this.authService.isLoggedIn();
  }

}


