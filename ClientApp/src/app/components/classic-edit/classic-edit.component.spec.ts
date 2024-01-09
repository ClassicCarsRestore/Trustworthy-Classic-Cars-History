import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClassicEditComponent } from './classic-edit.component';

describe('ClassicEditComponent', () => {
  let component: ClassicEditComponent;
  let fixture: ComponentFixture<ClassicEditComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClassicEditComponent]
    });
    fixture = TestBed.createComponent(ClassicEditComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
