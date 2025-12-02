import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from './auth';

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
vi.stubGlobal('localStorage', localStorageMock);

// Mock API
vi.mock('@/lib/api', () => ({
  api: {
    auth: {
      login: vi.fn().mockResolvedValue({
        token: 'test-token',
        user_id: 'test-user-id',
        store_id: 'test-store-id',
        store_name: 'Test Store',
      }),
      register: vi.fn().mockResolvedValue({
        token: 'test-token',
        user_id: 'test-user-id',
        store_id: 'test-store-id',
        store_name: 'Test Store',
      }),
    },
  },
}));

describe('Auth Store', () => {
  beforeEach(() => {
    // Reset the store before each test
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    });
    
    // Clear localStorage mocks
    vi.clearAllMocks();
  });

  it('initializes with empty state', () => {
    const { result } = renderHook(() => useAuthStore());
    
    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('logs in user successfully', async () => {
    const { result } = renderHook(() => useAuthStore());
    
    await act(async () => {
      await result.current.login('test@example.com', 'password');
    });
    
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.token).toBe('test-token');
    expect(result.current.user).toEqual({
      id: 'test-user-id',
      email: 'test@example.com',
      storeId: 'test-store-id',
      storeName: 'Test Store',
    });
    
    // Verify token is saved to localStorage
    expect(localStorageMock.setItem).toHaveBeenCalledWith(
      'auth_token',
      'test-token'
    );
  });

  it('registers user successfully', async () => {
    const { result } = renderHook(() => useAuthStore());
    
    await act(async () => {
      await result.current.register('test@example.com', 'password', 'Test Store');
    });
    
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.token).toBe('test-token');
    expect(result.current.user).toEqual({
      id: 'test-user-id',
      email: 'test@example.com',
      storeId: 'test-store-id',
      storeName: 'Test Store',
    });
  });

  it('logs out user successfully', async () => {
    const { result } = renderHook(() => useAuthStore());
    
    // First login
    await act(async () => {
      await result.current.login('test@example.com', 'password');
    });
    
    expect(result.current.isAuthenticated).toBe(true);
    
    // Then logout
    await act(async () => {
      result.current.logout();
    });
    
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.token).toBeNull();
    expect(result.current.user).toBeNull();
    
    // Verify token is removed from localStorage
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('auth_token');
  });

  it('fails login with invalid credentials', async () => {
    const { api } = await import('@/lib/api');
    vi.mocked(api.auth.login).mockRejectedValueOnce(
      new Error('Invalid credentials')
    );
    
    const { result } = renderHook(() => useAuthStore());
    
    await act(async () => {
      await expect(result.current.login('invalid@example.com', 'wrong'))
        .rejects.toThrow('Invalid credentials');
    });
    
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.token).toBeNull();
    expect(result.current.user).toBeNull();
  });

  it('initializes from localStorage when token exists', async () => {
    // Mock localStorage to have a token
    localStorageMock.getItem.mockReturnValue('existing-token');
    
    // Mock API to return user data
    const { api } = await import('@/lib/api');
    vi.mocked(api.auth.getCurrentUser).mockResolvedValueOnce({
      id: 'existing-user-id',
      email: 'existing@example.com',
      storeId: 'existing-store-id',
      storeName: 'Existing Store',
    });
    
    // Create a new store instance to test initialization
    const { result } = renderHook(() => useAuthStore());
    
    // Verify token was retrieved from localStorage
    expect(localStorageMock.getItem).toHaveBeenCalledWith('auth_token');
  });

  it('updates user store information', async () => {
    const { result } = renderHook(() => useAuthStore());
    
    // First login
    await act(async () => {
      await result.current.login('test@example.com', 'password');
    });
    
    expect(result.current.user?.storeName).toBe('Test Store');
    
    // Update store information
    await act(async () => {
      result.current.updateStore({
        storeName: 'Updated Store',
        subscriptionPlan: 'pro',
      });
    });
    
    expect(result.current.user?.storeName).toBe('Updated Store');
    expect(result.current.user?.subscriptionPlan).toBe('pro');
  });
});