export type NodeType = 'LEADER' | 'WORKER' | 'CLIENT';

export interface Event {
  id: string;
  type: 'tweet' | 'log' | 'transaction';
  content: string;
  timestamp: number; // Real-world arrival time
  lamportTime?: number;
  workerId?: string;
}

export interface Message {
  from: string;
  to: string;
  type: 'ASSIGN_TASK' | 'TASK_COMPLETED' | 'CLOCK_SYNC';
  payload: any;
  lamportTime: number;
}

export interface NodeState {
  id: string;
  type: NodeType;
  status: 'ACTIVE' | 'INACTIVE';
  clock: number;
}

export interface WorkerState extends NodeState {
  queueSize: number;
  processedCount: number;
}

export interface SimulationState {
  leader: NodeState;
  workers: WorkerState[];
  events: Event[];
  isRunning: boolean;
  pipelineLogs: string[];
  isProcessing: boolean;
  assignedWorkerId?: string;
}
