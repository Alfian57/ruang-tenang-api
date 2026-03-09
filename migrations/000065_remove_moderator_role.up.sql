-- Promote all moderator users to admin since moderator role is being removed
UPDATE users SET role = 'admin' WHERE role = 'moderator';
