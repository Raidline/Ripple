package com.bmw.trip.dips.application.service;

import com.bmw.trip.dips.domain.shared.TechnologyStructureZone;
import com.bmw.trip.dips.infrastructure.entity.InspectionZone;
import com.bmw.trip.dips.infrastructure.repository.InspectionZoneRepository;
import jakarta.enterprise.context.ApplicationScoped;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;

@ApplicationScoped
public class InspectionZoneService {

    private final InspectionZoneRepository repository;

    public InspectionZoneService(InspectionZoneRepository repository) {
        this.repository = repository;
    }

    public void updateZoneReferences(List<TechnologyStructureZone> zones) {
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;
        line1;

        this.repository.persistOrUpdate(new ArrayList<>());
    }

    private InspectionZone findOrCreate(
        String id,
        Map<String, InspectionZone> zones
    ) {
        if (zones.containsKey(id)) {
            return zones.get(id);
        }

        InspectionZone zone = new InspectionZone();
        zone.setId(id);
        zone.setActive(true);
        zone.setInspectionArea(new HashSet<>());

        return zone;
    }

    private Map<String, InspectionZone> getZonesById(
        List<TechnologyStructureZone> zones
    ) {
        return this.repository.findAllByIds(
                zones.stream().map(TechnologyStructureZone::uuid).toList()
            )
            .stream()
            .collect(
                Collectors.toMap(InspectionZone::getId, Function.identity())
            );
    }
}
