'use client';

import React, { useState, useEffect } from 'react';
import { LeftSidebar, RightSidebar } from '@/components/Sidebar';
import { Feed } from '@/components/Feed';
import { PipelinePanel } from '@/components/PipelinePanel';
import { simulationManager } from '@/lib/simulation';
import { SimulationState } from '@/lib/types';

export default function Home() {
  const [state, setState] = useState<SimulationState>(simulationManager.getState());

  // Periodically refresh state to reflect "processing"
  useEffect(() => {
    const interval = setInterval(() => {
      setState(simulationManager.getState());
    }, 100);
    return () => clearInterval(interval);
  }, []);

  const handleSimulate = (text: string) => {
    simulationManager.simulateClientEvent(text);
    setState(simulationManager.getState());
  };

  const handleReset = () => {
    simulationManager.reset();
    setState(simulationManager.getState());
  };

  return (
    <main className="min-h-screen bg-black flex justify-center">
      <div className="flex w-full max-w-[1240px]">
        {/* LEFT SIDEBAR */}
        <LeftSidebar
          state={state}
          onSimulate={handleSimulate}
          onReset={handleReset}
        />

        {/* CENTER COLUMN: FEED + PIPELINE */}
        <div className="flex-1 flex flex-col border-x border-gray-800 min-h-screen bg-black overflow-y-auto">
          <Feed events={state.events} />
          <PipelinePanel
            logs={state.pipelineLogs}
            assignedWorkerId={state.assignedWorkerId}
          />
        </div>

        {/* RIGHT SIDEBAR */}
        <RightSidebar state={state} />
      </div>
    </main>
  );
}
