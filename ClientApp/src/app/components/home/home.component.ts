import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  constructor(private router: Router, private authService: AuthService) {}

  ngOnInit() {
    if (this.authService.isLoggedIn()) {
      this.router.navigate(['/classics/']);
    } else {
      this.router.navigate(['/login/']);
    }
  }
}