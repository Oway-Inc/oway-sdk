import { describe, it, expect } from 'vitest';
import Oway from '../index';
import { Carrier } from '../resources/carrier';

describe('Oway client', () => {
  it('exposes a carrier resource', () => {
    const oway = new Oway({ clientId: 'id', clientSecret: 'secret' });
    expect(oway.carrier).toBeInstanceOf(Carrier);
  });
});
