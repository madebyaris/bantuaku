import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the login page
    await page.goto('/login');
  });

  test('should display login form correctly', async ({ page }) => {
    // Check if the login form is displayed
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Masuk' })).toBeVisible();
    await expect(page.getByText('Belum punya akun?')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Daftar' })).toBeVisible();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    // Fill in invalid credentials
    await page.getByLabel('Email').fill('invalid@example.com');
    await page.getByLabel('Password').fill('wrongpassword');
    
    // Submit the form
    await page.getByRole('button', { name: 'Masuk' }).click();
    
    // Check if error message is displayed
    await expect(page.getByText('Email atau password salah')).toBeVisible();
  });

  test('should navigate to registration page', async ({ page }) => {
    // Click the register link
    await page.getByRole('link', { name: 'Daftar' }).click();
    
    // Check if we're on the registration page
    await expect(page).toHaveURL('/register');
    await expect(page.getByLabel('Nama Toko')).toBeVisible();
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByLabel('Konfirmasi Password')).toBeVisible();
  });

  test('should register a new user successfully', async ({ page }) => {
    // Navigate to registration page
    await page.getByRole('link', { name: 'Daftar' }).click();
    
    // Fill in the registration form
    await page.getByLabel('Nama Toko').fill('Test Store');
    await page.getByLabel('Email').fill(`test-${Date.now()}@example.com`);
    await page.getByLabel('Password').fill('password123');
    await page.getByLabel('Konfirmasi Password').fill('password123');
    
    // Submit the form
    await page.getByRole('button', { name: 'Daftar' }).click();
    
    // Check if we're redirected to dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Check if user info is displayed
    await expect(page.getByText('Test Store')).toBeVisible();
  });

  test('should validate registration form', async ({ page }) => {
    // Navigate to registration page
    await page.getByRole('link', { name: 'Daftar' }).click();
    
    // Try to submit empty form
    await page.getByRole('button', { name: 'Daftar' }).click();
    
    // Check if validation errors are displayed
    await expect(page.getByText('Nama toko wajib diisi')).toBeVisible();
    await expect(page.getByText('Email wajib diisi')).toBeVisible();
    await expect(page.getByText('Password wajib diisi')).toBeVisible();
    
    // Test password confirmation mismatch
    await page.getByLabel('Nama Toko').fill('Test Store');
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password').fill('password123');
    await page.getByLabel('Konfirmasi Password').fill('differentpassword');
    await page.getByRole('button', { name: 'Daftar' }).click();
    
    await expect(page.getByText('Konfirmasi password tidak cocok')).toBeVisible();
  });

  test('should login with valid credentials', async ({ page }) => {
    // Fill in valid credentials (using the demo account)
    await page.getByLabel('Email').fill('demo@bantuaku.id');
    await page.getByLabel('Password').fill('demo123');
    
    // Submit the form
    await page.getByRole('button', { name: 'Masuk' }).click();
    
    // Check if we're redirected to dashboard
    await expect(page).toHaveURL('/dashboard');
    
    // Check if user info is displayed
    await expect(page.getByText('Toko Berkah Jaya')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.getByLabel('Email').fill('demo@bantuaku.id');
    await page.getByLabel('Password').fill('demo123');
    await page.getByRole('button', { name: 'Masuk' }).click();
    await expect(page).toHaveURL('/dashboard');
    
    // Logout
    await page.getByRole('button', { name: 'Keluar' }).click();
    
    // Check if we're redirected to login page
    await expect(page).toHaveURL('/login');
  });

  test('should protect routes when not authenticated', async ({ page }) => {
    // Try to access dashboard directly
    await page.goto('/dashboard');
    
    // Check if we're redirected to login page
    await expect(page).toHaveURL('/login');
  });
});