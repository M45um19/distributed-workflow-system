export interface IUserRepository {
  create(userData: any): Promise<any>;
  findByEmail(email: string): Promise<any>;
  exists(email: string): Promise<boolean | any>;
  findById(id: string): Promise<any>;
}