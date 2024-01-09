import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClassicDetailsComponent } from './classic-details.component';

describe('ClassicDetailsComponent', () => {
  let component: ClassicDetailsComponent;
  let fixture: ComponentFixture<ClassicDetailsComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClassicDetailsComponent]
    });
    fixture = TestBed.createComponent(ClassicDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
