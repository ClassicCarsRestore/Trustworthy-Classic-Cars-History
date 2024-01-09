import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClassicUpdateEmailComponent } from './classic-update-email.component';

describe('ClassicUpdateEmailComponent', () => {
  let component: ClassicUpdateEmailComponent;
  let fixture: ComponentFixture<ClassicUpdateEmailComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClassicUpdateEmailComponent]
    });
    fixture = TestBed.createComponent(ClassicUpdateEmailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
