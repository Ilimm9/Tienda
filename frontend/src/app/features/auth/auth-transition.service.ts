import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

export type AuthMode = 'login' | 'registro';

@Injectable({ providedIn: 'root' })
export class AuthTransitionService {
  private readonly modeRequests = new Subject<AuthMode>();

  readonly requests$ = this.modeRequests.asObservable();

  request(mode: AuthMode): void {
    this.modeRequests.next(mode);
  }
}
