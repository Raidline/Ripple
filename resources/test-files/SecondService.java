package com.bmw.trip.dips.application.service;

import java.util.List;
import com.bmw.trip.dips.infrastructure.service.InspectionZoneService;
import com.bmw.trip.dips.infrastructure.repository.InspectionZoneRepository;

@ApplicationScoped
public class SecondService {

    private final InspectionZoneService serv;
    private final InspectionZoneRepository repository;

    public InspectionZoneService(InspectionZoneService serv, InspectionZoneRepository repository) {
        this.serv = serv;
        this.repository = repository;
    }


    public String m1() {

        serv.updateZoneReferences(List.of());

        return "banana"
    }
}
