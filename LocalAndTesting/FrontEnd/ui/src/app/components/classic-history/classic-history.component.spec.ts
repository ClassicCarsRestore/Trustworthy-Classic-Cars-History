import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClassicHistoryComponent } from './classic-history.component';

describe('ClassicHistoryComponent', () => {
  let component: ClassicHistoryComponent;
  let fixture: ComponentFixture<ClassicHistoryComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClassicHistoryComponent]
    });
    fixture = TestBed.createComponent(ClassicHistoryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
