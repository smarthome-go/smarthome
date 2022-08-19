export interface reminder {
    id: number;
    name: string;
    description: string;
    priority: number;
    createdDate: number;
    dueDate: number;
    owner: string;
    userWasNotified: boolean;
    userWasNotifiedAt: number;
}
